package http

import (
	"log/slog"
	"net/http"
	"strings"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"

	"warehouse/internal/auth"
	"warehouse/internal/config"
	"warehouse/internal/domain"
	"warehouse/internal/repo"
	"warehouse/internal/service"
)

type Deps struct {
	Logger *slog.Logger
	DB     *repo.DB
	Cfg    config.Config
}

func NewRouter(d Deps) http.Handler {
	r := chi.NewRouter()

	r.Use(middleware.RequestID)
	r.Use(middleware.RealIP)
	r.Use(middleware.Recoverer)
	r.Use(middleware.Timeout(30 * time.Second))

	r.Get("/healthz", func(w http.ResponseWriter, r *http.Request) {
		JSON(w, http.StatusOK, map[string]any{"ok": true})
	})

	fileServer := http.FileServer(http.Dir("web"))
	r.Mount("/web", http.StripPrefix("/web", fileServer))

	r.Get("/", func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/web") {
			http.NotFound(w, r)
			return
		}
		http.Redirect(w, r, "/web/", http.StatusTemporaryRedirect)
	})

	jwtMgr := auth.NewManager(d.Cfg.JWTSecret)

	itemsRepo := repo.NewItemsRepo(d.DB)
	historyRepo := repo.NewHistoryRepo(d.DB)

	itemsSvc := service.NewItemsService(d.DB, itemsRepo)
	historySvc := service.NewHistoryService(historyRepo)

	itemsH := NewItemsHandler(itemsSvc)
	histH := NewHistoryHandler(historySvc)

	r.Route("/api", func(api chi.Router) {
		api.Get("/ping", func(w http.ResponseWriter, r *http.Request) {
			JSON(w, http.StatusOK, map[string]any{"pong": true})
		})

		api.Route("/auth", func(ar chi.Router) {
			ar.Post("/login", LoginHandler(jwtMgr))
		})

		api.Group(func(pr chi.Router) {
			pr.Use(RequireAuth(jwtMgr))

			pr.With(RequireRoles(domain.RoleViewer, domain.RoleManager, domain.RoleAdmin)).Get("/me", MeHandler())

			// items
			pr.Route("/items", func(ir chi.Router) {
				ir.With(RequireRoles(domain.RoleViewer, domain.RoleManager, domain.RoleAdmin)).Get("/", itemsH.List())
				ir.With(RequireRoles(domain.RoleManager, domain.RoleAdmin)).Post("/", itemsH.Create())
				ir.With(RequireRoles(domain.RoleManager, domain.RoleAdmin)).Put("/{id}", itemsH.Update())
				ir.With(RequireRoles(domain.RoleAdmin)).Delete("/{id}", itemsH.Delete())

				ir.With(RequireRoles(domain.RoleViewer, domain.RoleManager, domain.RoleAdmin)).Get("/{id}/history", histH.ListByItem())
				ir.With(RequireRoles(domain.RoleViewer, domain.RoleManager, domain.RoleAdmin)).Get("/{id}/history.csv", histH.ExportCSV())
			})
		})
	})

	return r
}
