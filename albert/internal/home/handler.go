package home

import (
	"log/slog"
	"net/http"

	"github.com/gorilla/csrf"

	"albert/internal/configs"
	"albert/internal/users"
	"albert/views"
)

type Handler struct {
	logger *slog.Logger
}

func NewHandler(logger *slog.Logger) *Handler {
	return &Handler{logger: logger}
}

func (h *Handler) ShowIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	params := views.IndexParams{
		CsrfName:    configs.CsrfFormFieldName,
		CsrfToken:   csrf.Token(r),
		CurrentUser: users.CurrentUserFromCtx(ctx),
	}

	if err := views.Index.Execute(w, params); err != nil {
		h.logger.Error("failed to render index template", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
