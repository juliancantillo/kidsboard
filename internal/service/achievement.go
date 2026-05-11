package service

import (
	"context"
	"fmt"

	"cantillo.dev/kidsboard/internal/domain"
	"cantillo.dev/kidsboard/internal/storage"
	"cantillo.dev/kidsboard/internal/storage/sqldb"
)

// AchievementService owns achievement evaluation. Reevaluate is the only
// load-bearing method: it runs inside the same transaction as a triggering
// event and returns the achievements newly earned in this call.
type AchievementService interface {
	Reevaluate(ctx context.Context, db storage.DBTX, kidID int64) ([]domain.Achievement, error)
}

func NewAchievementService(balance BalanceService) AchievementService {
	return &achievementService{balance: balance}
}

type achievementService struct {
	balance BalanceService
}

func (s *achievementService) Reevaluate(ctx context.Context, db storage.DBTX, kidID int64) ([]domain.Achievement, error) {
	q := sqldb.New(db)
	var newlyEarned []domain.Achievement
	for {
		rows, err := q.ListUnearnedAchievementRulesForKid(ctx, kidID)
		if err != nil {
			return nil, fmt.Errorf("list unearned: %w", err)
		}
		unearned := groupRulesByAchievement(rows)
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
		// Fixed-point: stop when a full pass yields no new earns. The unearned
		// list shrinks every iteration (an earned achievement is filtered by
		// the WHERE clause next pass), so termination is bounded by N.
		if passCount == 0 {
			return newlyEarned, nil
		}
	}
}

// groupRulesByAchievement folds the JOIN-flat rows from the engine query into
// one domain.Achievement per id, with rules attached in row order.
func groupRulesByAchievement(rows []sqldb.ListUnearnedAchievementRulesForKidRow) []domain.Achievement {
	byID := make(map[int64]*domain.Achievement)
	var order []int64
	for _, r := range rows {
		ach, ok := byID[r.AchievementID]
		if !ok {
			ach = &domain.Achievement{
				ID:          r.AchievementID,
				Slug:        r.AchievementSlug,
				Name:        r.AchievementName,
				Description: r.AchievementDescription,
				Title:       r.AchievementTitle,
				Combinator:  domain.Combinator(r.AchievementCombinator),
				BonusPoints: r.AchievementBonusPoints,
				ArchivedAt:  r.AchievementArchivedAt,
				CreatedAt:   r.AchievementCreatedAt,
			}
			byID[r.AchievementID] = ach
			order = append(order, r.AchievementID)
		}
		metric := domain.Metric(r.RuleMetric)
		threshold := r.RuleThreshold
		if metric == domain.MetricLevel {
			// Level is sugar for XP at the equivalent threshold. Transform
			// once at load so the evaluator only needs to know XP.
			metric = domain.MetricXP
			threshold = domain.XPForLevel(int(threshold))
		}
		ach.Rules = append(ach.Rules, domain.AchievementRule{
			ID:            r.RuleID,
			AchievementID: r.AchievementID,
			CategoryID:    r.RuleCategoryID,
			Metric:        metric,
			Threshold:     threshold,
		})
	}
	out := make([]domain.Achievement, 0, len(order))
	for _, id := range order {
		out = append(out, *byID[id])
	}
	return out
}

func (s *achievementService) evaluateRules(ctx context.Context, db storage.DBTX, kidID int64, ach domain.Achievement) (bool, error) {
	results := make([]bool, len(ach.Rules))
	for i, rule := range ach.Rules {
		current, err := s.computeMetric(ctx, db, kidID, rule)
		if err != nil {
			return false, err
		}
		results[i] = current >= rule.Threshold
	}
	return foldCombinator(ach.Combinator, results), nil
}

func (s *achievementService) computeMetric(ctx context.Context, db storage.DBTX, kidID int64, rule domain.AchievementRule) (int64, error) {
	q := sqldb.New(db)
	var categoryArg interface{}
	if rule.CategoryID != nil {
		categoryArg = *rule.CategoryID
	}
	switch rule.Metric {
	case domain.MetricCount:
		return q.CountActivitiesForKid(ctx, sqldb.CountActivitiesForKidParams{KidID: kidID, CategoryID: categoryArg})
	case domain.MetricXP:
		return q.SumActivityXPForKid(ctx, sqldb.SumActivityXPForKidParams{KidID: kidID, CategoryID: categoryArg})
	case domain.MetricPoints:
		if rule.CategoryID != nil {
			return s.balance.PointsEarnedInCategory(ctx, db, kidID, *rule.CategoryID)
		}
		return s.balance.Balance(ctx, db, kidID)
	}
	return 0, fmt.Errorf("unsupported metric %q", rule.Metric)
}

func foldCombinator(c domain.Combinator, results []bool) bool {
	switch c {
	case domain.CombinatorAll:
		for _, r := range results {
			if !r {
				return false
			}
		}
		return len(results) > 0
	case domain.CombinatorAny:
		for _, r := range results {
			if r {
				return true
			}
		}
	}
	return false
}
