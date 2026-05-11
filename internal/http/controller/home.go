package controller

import (
	"database/sql"
	"log"
	"net/http"

	"cantillo.dev/kidsboard/internal/domain"
	"cantillo.dev/kidsboard/internal/service"
	"cantillo.dev/kidsboard/internal/view"
)

type Home struct {
	db       *sql.DB
	renderer *view.Renderer
	kids     service.KidService
}

func NewHome(db *sql.DB, renderer *view.Renderer, kids service.KidService) *Home {
	return &Home{db: db, renderer: renderer, kids: kids}
}

type HomeView struct {
	Kids []domain.Kid
}

func (c *Home) Index(w http.ResponseWriter, r *http.Request) {
	kids, err := c.kids.ListActive(r.Context(), c.db)
	if err != nil {
		log.Printf("list kids: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
		return
	}
	if err := c.renderer.Render(w, "home", HomeView{Kids: kids}); err != nil {
		log.Printf("render home: %v", err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}
