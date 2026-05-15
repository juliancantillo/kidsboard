package controller

import (
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cantillo.dev/kidsboard/internal/domain"
	"cantillo.dev/kidsboard/internal/service"
	"cantillo.dev/kidsboard/internal/view"
)

// Admin serves the parent-only routes (no auth — trust the device).
type Admin struct {
	db            *sql.DB
	renderer      *view.Renderer
	kids          service.KidService
	categories    service.CategoryService
	activityTypes service.ActivityTypeService
	activities    service.ActivityService
}

func NewAdmin(
	db *sql.DB,
	renderer *view.Renderer,
	kids service.KidService,
	categories service.CategoryService,
	activityTypes service.ActivityTypeService,
	activities service.ActivityService,
) *Admin {
	return &Admin{
		db: db, renderer: renderer,
		kids: kids, categories: categories,
		activityTypes: activityTypes, activities: activities,
	}
}

type AdminIndexView struct {
	Kids          []domain.Kid // all (active + archived) — for the personajes grid
	ActiveKids    []domain.Kid // for the activity-log dropdown
	ActivityTypes []domain.ActivityType
	Categories    []domain.Category
}

func (c *Admin) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	kids, err := c.kids.ListAll(ctx, c.db)
	if err != nil {
		fail(w, "list kids", err)
		return
	}
	active := make([]domain.Kid, 0, len(kids))
	for _, k := range kids {
		if k.ArchivedAt == nil {
			active = append(active, k)
		}
	}
	types, err := c.activityTypes.ListActive(ctx, c.db)
	if err != nil {
		fail(w, "list activity types", err)
		return
	}
	cats, err := c.categories.ListActive(ctx, c.db)
	if err != nil {
		fail(w, "list categories", err)
		return
	}
	c.renderOrFail(w, "admin", AdminIndexView{
		Kids:          kids,
		ActiveKids:    active,
		ActivityTypes: types,
		Categories:    cats,
	})
}

type NewKidView struct {
	Avatars []view.Avatar
	Errors  map[string]string
	Form    NewKidForm
}

type NewKidForm struct {
	Name         string
	AvatarSlug   string
	Color        string
	DisplayOrder int
}

func (c *Admin) NewKid(w http.ResponseWriter, r *http.Request) {
	c.renderOrFail(w, "admin_kid_new", NewKidView{
		Avatars: view.Avatars,
		Form:    NewKidForm{Color: "#6366F1"},
	})
}

func (c *Admin) CreateKid(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	form := NewKidForm{
		Name:         strings.TrimSpace(r.FormValue("name")),
		AvatarSlug:   r.FormValue("avatar_slug"),
		Color:        strings.TrimSpace(r.FormValue("color")),
		DisplayOrder: parseInt(r.FormValue("display_order")),
	}
	_, err := c.kids.Create(r.Context(), c.db, service.CreateKidInput{
		Name: form.Name, AvatarSlug: form.AvatarSlug,
		Color: form.Color, DisplayOrder: form.DisplayOrder,
	})
	if err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			c.renderOrFail(w, "admin_kid_new", NewKidView{
				Avatars: view.Avatars, Errors: ve.Fields, Form: form,
			})
			return
		}
		fail(w, "create kid", err)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (c *Admin) ArchiveKid(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := c.kids.Archive(r.Context(), c.db, id); err != nil {
		fail(w, "archive kid", err)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (c *Admin) UnarchiveKid(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := c.kids.Unarchive(r.Context(), c.db, id); err != nil {
		fail(w, "unarchive kid", err)
		return
	}
	http.Redirect(w, r, "/admin", http.StatusSeeOther)
}

func (c *Admin) LogActivity(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	kidID, err := strconv.ParseInt(r.FormValue("kid_id"), 10, 64)
	if err != nil {
		http.Error(w, "bad kid_id", http.StatusBadRequest)
		return
	}
	atID, err := strconv.ParseInt(r.FormValue("activity_type_id"), 10, 64)
	if err != nil {
		http.Error(w, "bad activity_type_id", http.StatusBadRequest)
		return
	}
	qty := parseInt(r.FormValue("quantity"))
	if qty < 1 {
		qty = 1
	}

	if _, err := c.activities.Log(r.Context(), c.db, service.LogActivityInput{
		KidID:          kidID,
		ActivityTypeID: atID,
		Quantity:       qty,
		Note:           r.FormValue("note"),
	}); err != nil {
		fail(w, "log activity", err)
		return
	}
	http.Redirect(w, r, fmt.Sprintf("/kids/%d", kidID), http.StatusSeeOther)
}

func (c *Admin) renderOrFail(w http.ResponseWriter, page string, data any) {
	if err := c.renderer.Render(w, page, data); err != nil {
		log.Printf("render %q: %v", page, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func fail(w http.ResponseWriter, op string, err error) {
	log.Printf("%s: %v", op, err)
	http.Error(w, "internal error", http.StatusInternalServerError)
}

func parseInt(s string) int {
	n, err := strconv.Atoi(strings.TrimSpace(s))
	if err != nil {
		return 0
	}
	return n
}
