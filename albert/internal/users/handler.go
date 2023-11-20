package users

import (
	"context"
	"errors"
	"log/slog"
	"net/http"

	"github.com/alexedwards/scs/v2"
	"github.com/gorilla/csrf"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"albert/internal/configs"
	"albert/views"
)

type Handler struct {
	pool           *pgxpool.Pool
	sessionManager *scs.SessionManager
	logger         *slog.Logger
}

func NewHandler(pool *pgxpool.Pool, sessionManager *scs.SessionManager, logger *slog.Logger) *Handler {
	return &Handler{pool: pool, sessionManager: sessionManager, logger: logger}
}

func (h *Handler) ShowLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := views.LoginParams{
		CsrfName:  configs.CsrfFormFieldName,
		CsrfToken: csrf.Token(r),
	}
	if currentUser := CurrentUserFromCtx(ctx); currentUser != nil {
		http.Redirect(w, r, "/", http.StatusSeeOther)
		return
	}
	if err := views.Login.Execute(w, params); err != nil {
		h.logger.Error("failed to render login template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handler) DoLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	userId := r.PostFormValue("username")
	params := views.LoginParams{
		CsrfName:  configs.CsrfFormFieldName,
		CsrfToken: csrf.Token(r),
		Username:  userId,
	}

	user, err := h.findUserById(ctx, userId)
	if err != nil {
		h.logger.Error("failed to find user by id", "userId", userId, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if user == nil {
		params.ValidationError = "invalid username"
		if err = views.Login.Execute(w, params); err != nil {
			h.logger.Error("failed to render login template", "validationError", params.ValidationError)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}

	if err = h.sessionManager.RenewToken(ctx); err != nil {
		h.logger.Error("failed to renew session", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	h.sessionManager.Put(ctx, configs.SessionKeyUserId, user.Id)

	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) DoLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	h.sessionManager.Remove(ctx, configs.SessionKeyUserId)

	if err := h.sessionManager.RenewToken(ctx); err != nil {
		h.logger.Error("failed to renew session after logging out", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/auth/login", http.StatusSeeOther)
}

func (h *Handler) ShowWatchlist(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := views.WatchlistParams{
		CsrfName:    configs.CsrfFormFieldName,
		CsrfToken:   csrf.Token(r),
		CurrentUser: CurrentUserFromCtx(ctx),
	}

	var err error
	params.Watchlist, err = h.findUserFavorites(ctx, params.CurrentUser.Id)
	if err != nil {
		h.logger.Error("failed to find user favorites", "userId", params.CurrentUser.Id, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if err = views.Watchlist.Execute(w, params); err != nil {
		h.logger.Error("failed to render watchlist template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handler) Watch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser := CurrentUserFromCtx(ctx)
	ticker := r.PostFormValue("ticker")

	query := "insert into favorites (user_id, ticker) values ($1, $2);"
	if _, err := h.pool.Exec(ctx, query, currentUser.Id, ticker); err != nil {
		h.logger.Error("failed to add ticker to watchlist", "userId", currentUser.Id, "ticker", ticker, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) Unwatch(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser := CurrentUserFromCtx(ctx)
	ticker := r.PostFormValue("ticker")

	query := "delete from favorites where user_id = $1 and ticker = $2;"
	if _, err := h.pool.Exec(ctx, query, currentUser.Id, ticker); err != nil {
		h.logger.Error("failed to remove ticker from watchlist", "userId", currentUser.Id, "ticker", ticker, "error", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	http.Redirect(w, r, "/", http.StatusSeeOther)
}

func (h *Handler) findUserFavorites(ctx context.Context, userId string) ([]views.Security, error) {
	query := `
		select securities.ticker, securities.name, securities.last_price, securities.last_price_updated, true as watched
		from securities join favorites using (ticker)
		where favorites.user_id = $1;
	`
	rows, _ := h.pool.Query(ctx, query, userId)
	return pgx.CollectRows(rows, pgx.RowToStructByName[views.Security])
}

func (h *Handler) findUserById(ctx context.Context, userId string) (*views.User, error) {
	query := "select user_id, first_name, last_name, email from users where user_id = $1"
	row := h.pool.QueryRow(ctx, query, userId)

	var user views.User
	if err := row.Scan(&user.Id, &user.FirstName, &user.LastName, &user.Email); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, nil
		}
		h.logger.Error("unexpected error while finding user by id", "userId", userId, "error", err)
		return nil, err
	}
	return &user, nil
}

func (h *Handler) WithCurrentUser(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		ctx := request.Context()
		sessionId := h.sessionManager.GetString(ctx, configs.SessionKeyUserId)

		user, err := h.findUserById(ctx, sessionId)
		if err != nil {
			h.logger.Error("failed trying to find user by session id", "sessionId", sessionId, "error", err)
			next.ServeHTTP(writer, request)
			return
		}

		newCtx := context.WithValue(ctx, ctxKeyCurrentUser{}, user)
		next.ServeHTTP(writer, request.WithContext(newCtx))
	})
}

type ctxKeyCurrentUser struct{}

func CurrentUserFromCtx(ctx context.Context) *views.User {
	user, _ := ctx.Value(ctxKeyCurrentUser{}).(*views.User)
	return user
}

func RequireLogin(next http.Handler) http.Handler {
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if user := CurrentUserFromCtx(request.Context()); user != nil {
			next.ServeHTTP(writer, request)
			return
		}
		http.Redirect(writer, request, "/auth/login", http.StatusSeeOther)
	})
}
