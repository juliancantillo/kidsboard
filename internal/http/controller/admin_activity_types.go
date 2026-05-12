package controller

import (
	"database/sql"
	"errors"
	"log"
	"net/http"
	"strconv"
	"strings"

	"cantillo.dev/kidsboard/internal/domain"
	"cantillo.dev/kidsboard/internal/service"
	"cantillo.dev/kidsboard/internal/view"
)

// AdminActivityTypes wraps the activity-type CRUD parent flows. Activity
// types are the per-category templates (with XP + points config) that
// parents create and edit so they have something to log against.
type AdminActivityTypes struct {
	db            *sql.DB
	renderer      *view.Renderer
	categories    service.CategoryService
	activityTypes service.ActivityTypeService
}

func NewAdminActivityTypes(
	db *sql.DB,
	renderer *view.Renderer,
	categories service.CategoryService,
	activityTypes service.ActivityTypeService,
) *AdminActivityTypes {
	return &AdminActivityTypes{db: db, renderer: renderer, categories: categories, activityTypes: activityTypes}
}

// ActivityTypeGroup bundles all activity types under one category for the
// list page. The list view renders one section per group.
type ActivityTypeGroup struct {
	Category      domain.Category
	ActivityTypes []domain.ActivityType
}

type ActivityTypesListView struct {
	Groups []ActivityTypeGroup
}

func (c *AdminActivityTypes) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cats, err := c.categories.ListActive(ctx, c.db)
	if err != nil {
		fail(w, "list categories", err)
		return
	}
	types, err := c.activityTypes.ListAll(ctx, c.db)
	if err != nil {
		fail(w, "list activity types", err)
		return
	}
	byCategory := make(map[int64][]domain.ActivityType, len(cats))
	for _, t := range types {
		byCategory[t.CategoryID] = append(byCategory[t.CategoryID], t)
	}
	groups := make([]ActivityTypeGroup, 0, len(cats))
	for _, cat := range cats {
		groups = append(groups, ActivityTypeGroup{
			Category:      cat,
			ActivityTypes: byCategory[cat.ID],
		})
	}
	c.render(w, "admin_activity_types", ActivityTypesListView{Groups: groups})
}

// ActivityTypeFormView feeds both the New and Edit pages.
type ActivityTypeFormView struct {
	ActivityType domain.ActivityType // empty on create
	Form         ActivityTypeForm
	Categories   []domain.Category
	Errors       map[string]string
	IsEdit       bool
}

type ActivityTypeForm struct {
	CategoryID    int64
	Slug          string
	Name          string
	Description   string
	XPPerUnit     int64
	PointsPerUnit int64
}

func (c *AdminActivityTypes) New(w http.ResponseWriter, r *http.Request) {
	cats, err := c.categories.ListActive(r.Context(), c.db)
	if err != nil {
		fail(w, "list categories", err)
		return
	}
	form := ActivityTypeForm{
		XPPerUnit:     10,
		PointsPerUnit: 1,
	}
	// Pre-select category if ?category_id=<id> is passed (deep-link from list).
	if cid, _ := strconv.ParseInt(r.URL.Query().Get("category_id"), 10, 64); cid > 0 {
		form.CategoryID = cid
	}
	c.render(w, "admin_activity_type_form", ActivityTypeFormView{
		Form:       form,
		Categories: cats,
		IsEdit:     false,
	})
}

func (c *AdminActivityTypes) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	cats, err := c.categories.ListActive(r.Context(), c.db)
	if err != nil {
		fail(w, "list categories", err)
		return
	}
	form := parseActivityTypeForm(r)
	if _, err := c.activityTypes.Create(r.Context(), c.db, inputFromActivityTypeForm(form)); err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			c.render(w, "admin_activity_type_form", ActivityTypeFormView{
				Form: form, Categories: cats, Errors: ve.Fields, IsEdit: false,
			})
			return
		}
		fail(w, "create activity type", err)
		return
	}
	http.Redirect(w, r, "/admin/activity-types", http.StatusSeeOther)
}

func (c *AdminActivityTypes) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	at, err := c.activityTypes.Get(r.Context(), c.db, id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		fail(w, "get activity type", err)
		return
	}
	cats, err := c.categories.ListActive(r.Context(), c.db)
	if err != nil {
		fail(w, "list categories", err)
		return
	}
	c.render(w, "admin_activity_type_form", ActivityTypeFormView{
		ActivityType: at,
		Form:         formFromActivityType(at),
		Categories:   cats,
		IsEdit:       true,
	})
}

func (c *AdminActivityTypes) Update(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	cats, err := c.categories.ListActive(r.Context(), c.db)
	if err != nil {
		fail(w, "list categories", err)
		return
	}
	form := parseActivityTypeForm(r)
	if _, err := c.activityTypes.Update(r.Context(), c.db, id, inputFromActivityTypeForm(form)); err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			c.render(w, "admin_activity_type_form", ActivityTypeFormView{
				ActivityType: domain.ActivityType{ID: id},
				Form:         form, Categories: cats, Errors: ve.Fields, IsEdit: true,
			})
			return
		}
		fail(w, "update activity type", err)
		return
	}
	http.Redirect(w, r, "/admin/activity-types", http.StatusSeeOther)
}

func (c *AdminActivityTypes) Archive(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := c.activityTypes.Archive(r.Context(), c.db, id); err != nil {
		fail(w, "archive activity type", err)
		return
	}
	http.Redirect(w, r, "/admin/activity-types", http.StatusSeeOther)
}

func (c *AdminActivityTypes) Unarchive(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := c.activityTypes.Unarchive(r.Context(), c.db, id); err != nil {
		fail(w, "unarchive activity type", err)
		return
	}
	http.Redirect(w, r, "/admin/activity-types", http.StatusSeeOther)
}

func (c *AdminActivityTypes) render(w http.ResponseWriter, page string, data any) {
	if err := c.renderer.Render(w, page, data); err != nil {
		log.Printf("render %q: %v", page, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

func parseActivityTypeForm(r *http.Request) ActivityTypeForm {
	cid, _ := strconv.ParseInt(strings.TrimSpace(r.FormValue("category_id")), 10, 64)
	return ActivityTypeForm{
		CategoryID:    cid,
		Slug:          strings.TrimSpace(r.FormValue("slug")),
		Name:          strings.TrimSpace(r.FormValue("name")),
		Description:   r.FormValue("description"),
		XPPerUnit:     int64(parseInt(r.FormValue("xp_per_unit"))),
		PointsPerUnit: int64(parseInt(r.FormValue("points_per_unit"))),
	}
}

func inputFromActivityTypeForm(f ActivityTypeForm) service.ActivityTypeInput {
	return service.ActivityTypeInput{
		CategoryID:    f.CategoryID,
		Slug:          f.Slug,
		Name:          f.Name,
		Description:   f.Description,
		XPPerUnit:     f.XPPerUnit,
		PointsPerUnit: f.PointsPerUnit,
	}
}

func formFromActivityType(at domain.ActivityType) ActivityTypeForm {
	form := ActivityTypeForm{
		CategoryID:    at.CategoryID,
		Slug:          at.Slug,
		Name:          at.Name,
		XPPerUnit:     at.XPPerUnit,
		PointsPerUnit: at.PointsPerUnit,
	}
	if at.Description != nil {
		form.Description = *at.Description
	}
	return form
}
