package http

import (
	"io/fs"
	"net/http"

	"cantillo.dev/kidsboard/internal/http/controller"
	"cantillo.dev/kidsboard/internal/view"
)

// NewRouter wires every route. Uses Go 1.22+ pattern matching on http.ServeMux.
func NewRouter(deps Deps) http.Handler {
	mux := http.NewServeMux()

	staticFS, err := fs.Sub(view.Static, "static")
	if err != nil {
		panic(err) // boot-time, embedded FS is constant
	}
	mux.Handle("GET /static/", http.StripPrefix("/static/", http.FileServerFS(staticFS)))

	home := controller.NewHome(deps.DB, deps.Renderer, deps.Kids)
	profile := controller.NewProfile(deps.DB, deps.Renderer, deps.Profile)
	admin := controller.NewAdmin(deps.DB, deps.Renderer, deps.Kids, deps.Categories, deps.ActivityTypes, deps.Activities)
	adminAchievements := controller.NewAdminAchievements(deps.DB, deps.Renderer, deps.Categories, deps.Achievements)
	adminActivityTypes := controller.NewAdminActivityTypes(deps.DB, deps.Renderer, deps.Categories, deps.ActivityTypes)
	health := controller.NewHealth(deps.DB)

	mux.HandleFunc("GET /healthz", health.Live)
	mux.HandleFunc("GET /readyz", health.Ready)
	mux.HandleFunc("GET /{$}", home.Index)
	mux.HandleFunc("GET /kids/{id}", profile.Show)
	mux.HandleFunc("GET /admin", admin.Index)
	mux.HandleFunc("GET /admin/kids/new", admin.NewKid)
	mux.HandleFunc("POST /admin/kids", admin.CreateKid)
	mux.HandleFunc("POST /admin/kids/{id}/archive", admin.ArchiveKid)
	mux.HandleFunc("POST /admin/activities", admin.LogActivity)
	mux.HandleFunc("GET /admin/activity-types", adminActivityTypes.Index)
	mux.HandleFunc("GET /admin/activity-types/new", adminActivityTypes.New)
	mux.HandleFunc("POST /admin/activity-types", adminActivityTypes.Create)
	mux.HandleFunc("GET /admin/activity-types/{id}/edit", adminActivityTypes.Edit)
	mux.HandleFunc("POST /admin/activity-types/{id}", adminActivityTypes.Update)
	mux.HandleFunc("POST /admin/activity-types/{id}/archive", adminActivityTypes.Archive)
	mux.HandleFunc("POST /admin/activity-types/{id}/unarchive", adminActivityTypes.Unarchive)
	mux.HandleFunc("GET /admin/achievements", adminAchievements.Index)
	mux.HandleFunc("GET /admin/achievements/new", adminAchievements.New)
	mux.HandleFunc("POST /admin/achievements", adminAchievements.Create)
	mux.HandleFunc("GET /admin/achievements/{id}/edit", adminAchievements.Edit)
	mux.HandleFunc("POST /admin/achievements/{id}", adminAchievements.Update)
	mux.HandleFunc("POST /admin/achievements/{id}/archive", adminAchievements.Archive)

	return mux
}
