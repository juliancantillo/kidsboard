package service

import (
	"context"
	"fmt"

	"cantillo.dev/kidsboard/internal/domain"
	"cantillo.dev/kidsboard/internal/storage"
	"cantillo.dev/kidsboard/internal/storage/sqldb"
)

// GroupRulesByAchievement folds the flat (achievement,rule) rows from the
// engine query into one domain.Achievement per id, with rules attached in
// row order. The level→XP transformation is applied here so consumers only
// ever see count/xp/points metrics downstream.
func GroupRulesByAchievement(rows []sqldb.ListUnearnedAchievementRulesForKidRow) []domain.Achievement {
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

// ComputeRuleMetric reads the current value of a rule's metric for a kid.
// Used by both the engine (does it pass?) and progress views (how close?).
func ComputeRuleMetric(ctx context.Context, db storage.DBTX, kidID int64, rule domain.AchievementRule, balance BalanceService) (int64, error) {
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
			return balance.PointsEarnedInCategory(ctx, db, kidID, *rule.CategoryID)
		}
		return balance.Balance(ctx, db, kidID)
	}
	return 0, fmt.Errorf("unsupported metric %q", rule.Metric)
}

// FoldCombinator collapses N rule results into one boolean via ALL (AND) or
// ANY (OR). Empty rule list is never a pass under either combinator.
func FoldCombinator(c domain.Combinator, results []bool) bool {
	if len(results) == 0 {
		return false
	}
	switch c {
	case domain.CombinatorAll:
		for _, r := range results {
			if !r {
				return false
			}
		}
		return true
	case domain.CombinatorAny:
		for _, r := range results {
			if r {
				return true
			}
		}
	}
	return false
}

// CombinatorRatio returns the overall achievement-level progress ratio for a
// set of rule ratios. ALL uses min (bottleneck); ANY uses max (closest path).
func CombinatorRatio(c domain.Combinator, rules []domain.RuleProgress) float64 {
	if len(rules) == 0 {
		return 0
	}
	switch c {
	case domain.CombinatorAll:
		m := rules[0].Ratio
		for _, r := range rules[1:] {
			if r.Ratio < m {
				m = r.Ratio
			}
		}
		return m
	case domain.CombinatorAny:
		m := rules[0].Ratio
		for _, r := range rules[1:] {
			if r.Ratio > m {
				m = r.Ratio
			}
		}
		return m
	}
	return 0
}
