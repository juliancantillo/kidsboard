package controller

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"

	"cantillo.dev/kidsboard/internal/service"
	"cantillo.dev/kidsboard/internal/view"
)

type Profile struct {
	db       *sql.DB
	renderer *view.Renderer
	profile  service.ProfileService
}

func NewProfile(db *sql.DB, renderer *view.Renderer, profile service.ProfileService) *Profile {
	return &Profile{db: db, renderer: renderer, profile: profile}
}

func (c *Profile) Show(w http.ResponseWriter, r *http.Request) {
	idStr := r.PathValue("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		http.Error(w, "invalid id", http.StatusBadRequest)
		return
	}
	view, err := c.profile.BuildProfile(r.Context(), c.db, id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		log.Printf("build profile %d: %v", id, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := c.renderer.Render(w, "profile", view); err != nil {
		log.Printf("render profile: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
