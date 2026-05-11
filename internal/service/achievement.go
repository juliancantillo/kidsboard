package service

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"cantillo.dev/kidsboard/internal/domain"
	"cantillo.dev/kidsboard/internal/storage"
	"cantillo.dev/kidsboard/internal/storage/sqldb"
)

// AchievementService owns achievement CRUD AND evaluation. Reevaluate is the
// load-bearing engine method (runs inside a triggering event's tx). The CRUD
// methods power the parent admin pages.
type AchievementService interface {
	Reevaluate(ctx context.Context, db storage.DBTX, kidID int64) ([]domain.Achievement, error)

	ListAllWithRules(ctx context.Context, db storage.DBTX) ([]domain.Achievement, error)
	GetWithRules(ctx context.Context, db storage.DBTX, id int64) (domain.Achievement, error)
	Create(ctx context.Context, db *sql.DB, in AchievementInput) (domain.Achievement, error)
	Update(ctx context.Context, db *sql.DB, id int64, in AchievementInput) (domain.Achievement, error)
	Archive(ctx context.Context, db storage.DBTX, id int64) error
	Unarchive(ctx context.Context, db storage.DBTX, id int64) error
}

// AchievementInput is the parent-facing form shape for create/update.
// Title and Description empty-string == NULL in DB. Rules are replaced
// wholesale on update — the in-code spec is authoritative.
type AchievementInput struct {
	Slug        string
	Name        string
	Description string
	Title       string
	Combinator  string // ALL | ANY
	BonusPoints int64
	Rules       []AchievementRuleInput
}

type AchievementRuleInput struct {
	CategoryID *int64 // nil = global
	Metric     string // count | xp | points | level
	Threshold  int64
}

func NewAchievementService(balance BalanceService) AchievementService {
	return &achievementService{balance: balance}
}

type achievementService struct {
	balance BalanceService
}

// -- Engine -----------------------------------------------------------------

func (s *achievementService) Reevaluate(ctx context.Context, db storage.DBTX, kidID int64) ([]domain.Achievement, error) {
	q := sqldb.New(db)
	var newlyEarned []domain.Achievement
	for {
		rows, err := q.ListUnearnedAchievementRulesForKid(ctx, kidID)
		if err != nil {
			return nil, fmt.Errorf("list unearned: %w", err)
		}
		unearned := GroupRulesByAchievement(rows)
		var passCount int
		for _, ach := range unearned {
			passes, err := s.evaluateRules(ctx, db, kidID, ach)
			if err != nil {
				return nil, err
			}
			if !passes {
				continue
			}
			if err := q.MarkAchievementEarned(ctx, sqldb.MarkAchievementEarnedParams{
				KidID: kidID, AchievementID: ach.ID,
			}); err != nil {
				return nil, fmt.Errorf("mark earned: %w", err)
			}
			newlyEarned = append(newlyEarned, ach)
			passCount++
		}
		if passCount == 0 {
			return newlyEarned, nil
		}
	}
}

func (s *achievementService) evaluateRules(ctx context.Context, db storage.DBTX, kidID int64, ach domain.Achievement) (bool, error) {
	results := make([]bool, len(ach.Rules))
	for i, rule := range ach.Rules {
		current, err := ComputeRuleMetric(ctx, db, kidID, rule, s.balance)
		if err != nil {
			return false, err
		}
		results[i] = current >= rule.Threshold
	}
	return FoldCombinator(ach.Combinator, results), nil
}

// -- CRUD -------------------------------------------------------------------

func (s *achievementService) ListAllWithRules(ctx context.Context, db storage.DBTX) ([]domain.Achievement, error) {
	q := sqldb.New(db)
	// Admin sees archived too — sorted active-first, then archived dimmed.
	rows, err := q.ListAllAchievements(ctx)
	if err != nil {
		return nil, fmt.Errorf("list achievements: %w", err)
	}
	out := make([]domain.Achievement, 0, len(rows))
	for _, r := range rows {
		ach := achievementFromRow(r)
		rules, err := q.ListAchievementRules(ctx, r.ID)
		if err != nil {
			return nil, fmt.Errorf("list rules for %d: %w", r.ID, err)
		}
		ach.Rules = make([]domain.AchievementRule, 0, len(rules))
		for _, rr := range rules {
			ach.Rules = append(ach.Rules, ruleFromRow(rr))
		}
		out = append(out, ach)
	}
	return out, nil
}

func (s *achievementService) GetWithRules(ctx context.Context, db storage.DBTX, id int64) (domain.Achievement, error) {
	q := sqldb.New(db)
	row, err := q.GetAchievement(ctx, id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.Achievement{}, ErrNotFound
		}
		return domain.Achievement{}, fmt.Errorf("get achievement: %w", err)
	}
	ach := achievementFromRow(row)
	rules, err := q.ListAchievementRules(ctx, id)
	if err != nil {
		return domain.Achievement{}, fmt.Errorf("list rules: %w", err)
	}
	ach.Rules = make([]domain.AchievementRule, 0, len(rules))
	for _, r := range rules {
		ach.Rules = append(ach.Rules, ruleFromRow(r))
	}
	return ach, nil
}

func (s *achievementService) Create(ctx context.Context, db *sql.DB, in AchievementInput) (domain.Achievement, error) {
	if err := s.validate(in); err != nil {
		return domain.Achievement{}, err
	}
	return storage.WithTx(ctx, db, func(tx storage.DBTX) (domain.Achievement, error) {
		q := sqldb.New(tx)
		row, err := q.CreateAchievement(ctx, sqldb.CreateAchievementParams{
			Slug:        in.Slug,
			Name:        in.Name,
			Description: optString(in.Description),
			Title:       optString(in.Title),
			Combinator:  in.Combinator,
			BonusPoints: in.BonusPoints,
		})
		if err != nil {
			return domain.Achievement{}, fmt.Errorf("insert achievement: %w", err)
		}
		if err := insertRules(ctx, q, row.ID, in.Rules); err != nil {
			return domain.Achievement{}, err
		}
		ach := achievementFromRow(row)
		ach.Rules = ruleInputsToDomain(row.ID, in.Rules)
		return ach, nil
	})
}

func (s *achievementService) Update(ctx context.Context, db *sql.DB, id int64, in AchievementInput) (domain.Achievement, error) {
	if err := s.validate(in); err != nil {
		return domain.Achievement{}, err
	}
	return storage.WithTx(ctx, db, func(tx storage.DBTX) (domain.Achievement, error) {
		q := sqldb.New(tx)
		if err := q.UpdateAchievement(ctx, sqldb.UpdateAchievementParams{
			ID:          id,
			Name:        in.Name,
			Description: optString(in.Description),
			Title:       optString(in.Title),
			Combinator:  in.Combinator,
			BonusPoints: in.BonusPoints,
		}); err != nil {
			return domain.Achievement{}, fmt.Errorf("update achievement: %w", err)
		}
		// Rules are replaced wholesale — the form is authoritative.
		if err := q.DeleteAchievementRules(ctx, id); err != nil {
			return domain.Achievement{}, fmt.Errorf("clear rules: %w", err)
		}
		if err := insertRules(ctx, q, id, in.Rules); err != nil {
			return domain.Achievement{}, err
		}
		row, err := q.GetAchievement(ctx, id)
		if err != nil {
			return domain.Achievement{}, fmt.Errorf("reload achievement: %w", err)
		}
		ach := achievementFromRow(row)
		ach.Rules = ruleInputsToDomain(id, in.Rules)
		return ach, nil
	})
}

func (s *achievementService) Archive(ctx context.Context, db storage.DBTX, id int64) error {
	return sqldb.New(db).ArchiveAchievement(ctx, id)
}

func (s *achievementService) Unarchive(ctx context.Context, db storage.DBTX, id int64) error {
	return sqldb.New(db).UnarchiveAchievement(ctx, id)
}

// -- Validation -------------------------------------------------------------

func (s *achievementService) validate(in AchievementInput) error {
	fields := map[string]string{}
	if !isValidSlug(in.Slug) {
		fields["slug"] = "Slug requerido (a-z, 0-9, guiones; 1-60 caracteres)."
	}
	if name := strings.TrimSpace(in.Name); name == "" || len(name) > 100 {
		fields["name"] = "El nombre debe tener entre 1 y 100 caracteres."
	}
	if in.Combinator != string(domain.CombinatorAll) && in.Combinator != string(domain.CombinatorAny) {
		fields["combinator"] = "Combinador inválido (ALL o ANY)."
	}
	if in.BonusPoints < 0 {
		fields["bonus_points"] = "Los puntos extra no pueden ser negativos."
	}
	if len(in.Rules) == 0 {
		fields["rules"] = "Se requiere al menos una regla."
	}
	validMetrics := map[string]struct{}{"count": {}, "xp": {}, "points": {}, "level": {}}
	for i, r := range in.Rules {
		key := fmt.Sprintf("rule_%d", i)
		if _, ok := validMetrics[r.Metric]; !ok {
			fields[key+"_metric"] = "Métrica inválida."
		}
		if r.Threshold <= 0 {
			fields[key+"_threshold"] = "El umbral debe ser mayor que 0."
		}
	}
	if len(fields) > 0 {
		return &ValidationError{Fields: fields}
	}
	return nil
}

func isValidSlug(s string) bool {
	s = strings.TrimSpace(s)
	if len(s) < 1 || len(s) > 60 {
		return false
	}
	for _, c := range s {
		ok := (c >= 'a' && c <= 'z') || (c >= '0' && c <= '9') || c == '-'
		if !ok {
			return false
		}
	}
	return true
}

// -- Helpers ----------------------------------------------------------------

func insertRules(ctx context.Context, q *sqldb.Queries, achievementID int64, rules []AchievementRuleInput) error {
	for _, r := range rules {
		if _, err := q.InsertAchievementRule(ctx, sqldb.InsertAchievementRuleParams{
			AchievementID: achievementID,
			CategoryID:    r.CategoryID,
			Metric:        r.Metric,
			Threshold:     r.Threshold,
		}); err != nil {
			return fmt.Errorf("insert rule: %w", err)
		}
	}
	return nil
}

func ruleInputsToDomain(achievementID int64, in []AchievementRuleInput) []domain.AchievementRule {
	out := make([]domain.AchievementRule, 0, len(in))
	for _, r := range in {
		out = append(out, domain.AchievementRule{
			AchievementID: achievementID,
			CategoryID:    r.CategoryID,
			Metric:        domain.Metric(r.Metric),
			Threshold:     r.Threshold,
		})
	}
	return out
}

func achievementFromRow(r sqldb.Achievement) domain.Achievement {
	return domain.Achievement{
		ID: r.ID, Slug: r.Slug, Name: r.Name,
		Description: r.Description, Title: r.Title,
		Combinator:  domain.Combinator(r.Combinator),
		BonusPoints: r.BonusPoints, ArchivedAt: r.ArchivedAt, CreatedAt: r.CreatedAt,
	}
}

func ruleFromRow(r sqldb.AchievementRule) domain.AchievementRule {
	return domain.AchievementRule{
		ID: r.ID, AchievementID: r.AchievementID,
		CategoryID: r.CategoryID,
		Metric:     domain.Metric(r.Metric),
		Threshold:  r.Threshold,
	}
}

func optString(s string) *string {
	s = strings.TrimSpace(s)
	if s == "" {
		return nil
	}
	return &s
}
