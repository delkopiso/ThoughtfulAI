package securities

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/Masterminds/squirrel"
	"github.com/gorilla/csrf"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"albert/internal/configs"
	"albert/internal/users"
	"albert/views"
)

const baseUrl = "https://app.albert.com/casestudy/stock"

type Handler struct {
	apiKey string
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewHandler(apiKey string, pool *pgxpool.Pool, logger *slog.Logger) *Handler {
	return &Handler{apiKey: apiKey, pool: pool, logger: logger}
}

func (h *Handler) Search(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	currentUser := users.CurrentUserFromCtx(ctx)
	searchTerm := strings.TrimSpace(r.PostFormValue("search"))
	if len(searchTerm) == 0 {
		params := views.SearchSecuritiesParams{
			CsrfName:  configs.CsrfFormFieldName,
			CsrfToken: csrf.Token(r),
			Message:   "Type a few characters to start searching",
		}
		if err := views.SearchSecurities.Execute(w, params); err != nil {
			h.logger.Error("failed to render search securities partial", "error", err)
			w.WriteHeader(http.StatusInternalServerError)
		}
		return
	}
	query := `
			select securities.ticker, securities.name, securities.last_price, securities.last_price_updated
			     , case when favorites.user_id is null then false else true end as watched
			from securities
			left join favorites on securities.ticker = favorites.ticker and favorites.user_id = $1
			where securities.ticker ilike '%' || $2 || '%'
			or securities.name ilike '%' || $2 || '%';
		`
	rows, _ := h.pool.Query(ctx, query, currentUser.Id, searchTerm)
	securities, err := pgx.CollectRows(rows, pgx.RowToStructByName[views.Security])
	if err != nil {
		h.logger.Error("query to search securities failed", slog.String("error", err.Error()))
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	params := views.SearchSecuritiesParams{
		CsrfName:   configs.CsrfFormFieldName,
		CsrfToken:  csrf.Token(r),
		Securities: securities,
	}
	if err = views.SearchSecurities.Execute(w, params); err != nil {
		h.logger.Error("failed to render search securities partial", "error", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (h *Handler) LoadSecurities(ctx context.Context, frequency time.Duration) {
	ticker := time.NewTicker(frequency)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("stopping securities loader due to context cancellation")
			return
		case t := <-ticker.C:
			h.logger.Info("running load securities operation", slog.Time("time", t))
			if err := h.fetchAndLoadSecurities(ctx); err != nil {
				h.logger.Error("fetching and loading securities failed", slog.Any("error", err))
			}
		}
	}
}

func (h *Handler) LoadPrices(ctx context.Context, frequency, maxAge time.Duration) {
	ticker := time.NewTicker(frequency)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			h.logger.Info("stopping prices loader due to context cancellation")
			return
		case t := <-ticker.C:
			h.logger.Info("running load prices operation", slog.Time("time", t))
			if err := h.fetchAndLoadPrices(ctx, maxAge); err != nil {
				h.logger.Error("fetching and loading prices failed", slog.Any("error", err))
			}
		}
	}
}

func (h *Handler) doGet(ctx context.Context, url string, params url.Values) (*http.Response, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to fetch securities: %w", err)
	}

	request.Header.Set("Albert-Case-Study-API-Key", h.apiKey)
	request.URL.RawQuery = params.Encode()
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return nil, fmt.Errorf("failed to perform GET request on %s: %w", request.URL.String(), err)
	}
	if response.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received an unexpected status from %s: %d", request.URL.String(), response.StatusCode)
	}
	return response, nil
}

func (h *Handler) fetchAndLoadSecurities(ctx context.Context) error {
	response, err := h.doGet(ctx, baseUrl+"/tickers/", nil)
	if err != nil {
		return err
	}
	m := map[string]string{}
	if err = json.NewDecoder(response.Body).Decode(&m); err != nil {
		return fmt.Errorf("failed to unmarshal fetch securities response from JSON: %w", err)
	}

	builder := squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar).Insert("securities")
	builder = builder.Columns("ticker", "name")
	for ticker, name := range m {
		builder = builder.Values(ticker, name)
	}
	builder = builder.Suffix("on conflict do nothing")
	sql, args, err := builder.ToSql()
	if err != nil {
		return fmt.Errorf("failed to generate sql query to upsert securities: %w", err)
	}

	_, err = h.pool.Exec(ctx, sql, args...)
	return err
}

func (h *Handler) fetchAndLoadPrices(ctx context.Context, age time.Duration) error {
	stalePricesQuery := `
		select distinct ticker from securities where last_price_updated is null or last_price_updated < $1
	`
	cutoffTime := time.Now().UTC().Add(-1 * age)
	rows, err := h.pool.Query(ctx, stalePricesQuery, cutoffTime)
	if err != nil {
		return fmt.Errorf("query for most stale prices failed: %w", err)
	}
	defer rows.Close()

	tickers, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		return fmt.Errorf("failed to parse query most stale prices query results: %w", err)
	}

	if len(tickers) == 0 {
		h.logger.Info("found no stale prices, exiting early")
		return nil
	}

	params := url.Values{}
	params.Set("tickers", strings.Join(tickers, ","))
	response, err := h.doGet(ctx, baseUrl+"/prices/", params)
	if err != nil {
		return err
	}

	m := map[string]float64{}
	if err = json.NewDecoder(response.Body).Decode(&m); err != nil {
		return fmt.Errorf("failed to unmarshal fetch prices response from JSON: %w", err)
	}
	tickers = make([]string, 0, len(m))
	prices := make([]float64, 0, len(m))
	for ticker, price := range m {
		tickers = append(tickers, ticker)
		prices = append(prices, price)
	}

	updatePricesQuery := `
		update securities
		set last_price = updates.price, last_price_updated = now()
		from (select unnest($1::text[]) as ticker, unnest($2::numeric[]) as price) as updates
		where securities.ticker = updates.ticker;
	`

	_, err = h.pool.Exec(ctx, updatePricesQuery, tickers, prices)
	return err
}
