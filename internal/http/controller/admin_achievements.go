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

// AdminAchievements wraps the achievement-CRUD parent flows. Lives in its
// own struct so the main Admin controller stays focused on kids/activities.
type AdminAchievements struct {
	db           *sql.DB
	renderer     *view.Renderer
	categories   service.CategoryService
	achievements service.AchievementService
}

func NewAdminAchievements(
	db *sql.DB,
	renderer *view.Renderer,
	categories service.CategoryService,
	achievements service.AchievementService,
) *AdminAchievements {
	return &AdminAchievements{db: db, renderer: renderer, categories: categories, achievements: achievements}
}

type AchievementsListView struct {
	Achievements []domain.Achievement
	Categories   []domain.Category
}

func (c *AdminAchievements) Index(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	achs, err := c.achievements.ListAllWithRules(ctx, c.db)
	if err != nil {
		fail(w, "list achievements", err)
		return
	}
	cats, err := c.categories.ListActive(ctx, c.db)
	if err != nil {
		fail(w, "list categories", err)
		return
	}
	c.render(w, "admin_achievements", AchievementsListView{Achievements: achs, Categories: cats})
}

// AchievementFormView feeds both the New and Edit pages. Editing sets
// Achievement (and a non-zero ID); creating leaves it zero.
type AchievementFormView struct {
	Achievement domain.Achievement // empty on create
	Form        AchievementForm
	Categories  []domain.Category
	Errors      map[string]string
	IsEdit      bool
}

type AchievementForm struct {
	Slug        string
	Name        string
	Description string
	Title       string
	Combinator  string
	BonusPoints int64
	Rules       []AchievementRuleForm // pre-padded to >=minRuleSlots for the form
}

const minRuleSlots = 4

// padRules grows the slice so the template can render `minRuleSlots` rule
// rows without out-of-range index issues. Empty rows are stripped server-side
// before validation.
func padRules(rules []AchievementRuleForm) []AchievementRuleForm {
	for len(rules) < minRuleSlots {
		rules = append(rules, AchievementRuleForm{})
	}
	return rules
}

type AchievementRuleForm struct {
	CategorySlug string // "" = global
	Metric       string
	Threshold    int64
}

func (c *AdminAchievements) New(w http.ResponseWriter, r *http.Request) {
	cats, err := c.categories.ListActive(r.Context(), c.db)
	if err != nil {
		fail(w, "list categories", err)
		return
	}
	c.render(w, "admin_achievement_form", AchievementFormView{
		Form: AchievementForm{
			Combinator: "ALL",
			Rules:      padRules([]AchievementRuleForm{{Metric: "count", Threshold: 10}}),
		},
		Categories: cats,
		IsEdit:     false,
	})
}

func (c *AdminAchievements) Create(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	cats, err := c.categories.ListActive(r.Context(), c.db)
	if err != nil {
		fail(w, "list categories", err)
		return
	}
	form := parseAchievementForm(r)
	input, err := buildInput(form, cats)
	if err != nil {
		form.Rules = padRules(form.Rules)
		c.render(w, "admin_achievement_form", AchievementFormView{
			Form: form, Categories: cats, IsEdit: false,
			Errors: map[string]string{"rules": err.Error()},
		})
		return
	}

	if _, err := c.achievements.Create(r.Context(), c.db, input); err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			form.Rules = padRules(form.Rules)
			c.render(w, "admin_achievement_form", AchievementFormView{
				Form: form, Categories: cats, Errors: ve.Fields, IsEdit: false,
			})
			return
		}
		fail(w, "create achievement", err)
		return
	}
	http.Redirect(w, r, "/admin/achievements", http.StatusSeeOther)
}

func (c *AdminAchievements) Edit(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	ach, err := c.achievements.GetWithRules(r.Context(), c.db, id)
	if err != nil {
		if errors.Is(err, service.ErrNotFound) {
			http.NotFound(w, r)
			return
		}
		fail(w, "get achievement", err)
		return
	}
	cats, err := c.categories.ListActive(r.Context(), c.db)
	if err != nil {
		fail(w, "list categories", err)
		return
	}
	form := formFromAchievement(ach, cats)
	form.Rules = padRules(form.Rules)
	c.render(w, "admin_achievement_form", AchievementFormView{
		Achievement: ach,
		Form:        form,
		Categories:  cats,
		IsEdit:      true,
	})
}

func (c *AdminAchievements) Update(w http.ResponseWriter, r *http.Request) {
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
	form := parseAchievementForm(r)
	input, err := buildInput(form, cats)
	if err != nil {
		form.Rules = padRules(form.Rules)
		c.render(w, "admin_achievement_form", AchievementFormView{
			Achievement: domain.Achievement{ID: id},
			Form:        form, Categories: cats, IsEdit: true,
			Errors: map[string]string{"rules": err.Error()},
		})
		return
	}
	if _, err := c.achievements.Update(r.Context(), c.db, id, input); err != nil {
		var ve *service.ValidationError
		if errors.As(err, &ve) {
			form.Rules = padRules(form.Rules)
			c.render(w, "admin_achievement_form", AchievementFormView{
				Achievement: domain.Achievement{ID: id},
				Form:        form, Categories: cats, Errors: ve.Fields, IsEdit: true,
			})
			return
		}
		fail(w, "update achievement", err)
		return
	}
	http.Redirect(w, r, "/admin/achievements", http.StatusSeeOther)
}

func (c *AdminAchievements) Archive(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.ParseInt(r.PathValue("id"), 10, 64)
	if err != nil {
		http.Error(w, "bad id", http.StatusBadRequest)
		return
	}
	if err := c.achievements.Archive(r.Context(), c.db, id); err != nil {
		fail(w, "archive achievement", err)
		return
	}
	http.Redirect(w, r, "/admin/achievements", http.StatusSeeOther)
}

func (c *AdminAchievements) render(w http.ResponseWriter, page string, data any) {
	if err := c.renderer.Render(w, page, data); err != nil {
		log.Printf("render %q: %v", page, err)
		http.Error(w, "internal error", http.StatusInternalServerError)
	}
}

// parseAchievementForm reads the parallel-array rule encoding from the form.
// Empty rule slots (no metric selected) are dropped — the form preallocates
// extra slots for ergonomic editing.
func parseAchievementForm(r *http.Request) AchievementForm {
	form := AchievementForm{
		Slug:        strings.TrimSpace(r.FormValue("slug")),
		Name:        strings.TrimSpace(r.FormValue("name")),
		Description: r.FormValue("description"),
		Title:       r.FormValue("title"),
		Combinator:  strings.TrimSpace(r.FormValue("combinator")),
		BonusPoints: int64(parseInt(r.FormValue("bonus_points"))),
	}

	cats := r.Form["rule_category_slug"]
	metrics := r.Form["rule_metric"]
	thresholds := r.Form["rule_threshold"]
	n := max3(len(cats), len(metrics), len(thresholds))
	for i := 0; i < n; i++ {
		var rule AchievementRuleForm
		if i < len(cats) {
			rule.CategorySlug = strings.TrimSpace(cats[i])
		}
		if i < len(metrics) {
			rule.Metric = strings.TrimSpace(metrics[i])
		}
		if i < len(thresholds) {
			rule.Threshold = int64(parseInt(thresholds[i]))
		}
		// Skip wholly-empty rows so blank trailing slots don't fail validation.
		if rule.Metric == "" && rule.Threshold == 0 && rule.CategorySlug == "" {
			continue
		}
		form.Rules = append(form.Rules, rule)
	}
	return form
}

func buildInput(form AchievementForm, cats []domain.Category) (service.AchievementInput, error) {
	bySlug := make(map[string]int64, len(cats))
	for _, c := range cats {
		bySlug[c.Slug] = c.ID
	}
	rules := make([]service.AchievementRuleInput, 0, len(form.Rules))
	for _, r := range form.Rules {
		ruleInput := service.AchievementRuleInput{
			Metric:    r.Metric,
			Threshold: r.Threshold,
		}
		if r.CategorySlug != "" {
			id, ok := bySlug[r.CategorySlug]
			if !ok {
				return service.AchievementInput{}, errors.New("Una regla referencia una categoría desconocida.")
			}
			ruleInput.CategoryID = &id
		}
		rules = append(rules, ruleInput)
	}
	return service.AchievementInput{
		Slug:        form.Slug,
		Name:        form.Name,
		Description: form.Description,
		Title:       form.Title,
		Combinator:  form.Combinator,
		BonusPoints: form.BonusPoints,
		Rules:       rules,
	}, nil
}

func formFromAchievement(ach domain.Achievement, cats []domain.Category) AchievementForm {
	byID := make(map[int64]string, len(cats))
	for _, c := range cats {
		byID[c.ID] = c.Slug
	}
	form := AchievementForm{
		Slug:        ach.Slug,
		Name:        ach.Name,
		Combinator:  string(ach.Combinator),
		BonusPoints: ach.BonusPoints,
	}
	if ach.Description != nil {
		form.Description = *ach.Description
	}
	if ach.Title != nil {
		form.Title = *ach.Title
	}
	for _, r := range ach.Rules {
		rf := AchievementRuleForm{
			Metric:    string(r.Metric),
			Threshold: r.Threshold,
		}
		if r.CategoryID != nil {
			rf.CategorySlug = byID[*r.CategoryID]
		}
		form.Rules = append(form.Rules, rf)
	}
	return form
}

func max3(a, b, c int) int {
	m := a
	if b > m {
		m = b
	}
	if c > m {
		m = c
	}
	return m
}
