package service

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"cantillo.dev/kidsboard/internal/domain"
	"cantillo.dev/kidsboard/internal/storage"
	"cantillo.dev/kidsboard/internal/storage/sqldb"
)

// ProfileService composes the per-kid profile read model from many sources.
// Read-only — never mutates. The HTTP layer renders templates directly off
// the returned ProfileView.
type ProfileService interface {
	BuildProfile(ctx context.Context, db storage.DBTX, kidID int64) (domain.ProfileView, error)
}

func NewProfileService(balance BalanceService) ProfileService {
	return &profileService{balance: balance}
}

type profileService struct {
	balance BalanceService
}

const (
	recentActivityLimit  = 10
	nextAchievementLimit = 3
)

func (s *profileService) BuildProfile(ctx context.Context, db storage.DBTX, kidID int64) (domain.ProfileView, error) {
	q := sqldb.New(db)

	kidRow, err := q.GetKid(ctx, kidID)
	if err != nil {
		if errors.Is(err, errNoRows()) {
			return domain.ProfileView{}, ErrNotFound
		}
		return domain.ProfileView{}, fmt.Errorf("get kid: %w", err)
	}
	kid := domain.Kid{
		ID: kidRow.ID, Name: kidRow.Name, AvatarSlug: kidRow.AvatarSlug,
		Color: kidRow.Color, DisplayOrder: int(kidRow.DisplayOrder),
		ArchivedAt: kidRow.ArchivedAt, CreatedAt: kidRow.CreatedAt,
	}

	balance, err := s.composeBalance(ctx, db, kidID)
	if err != nil {
		return domain.ProfileView{}, err
	}

	catLevels, err := s.composeCategoryLevels(ctx, db, q, kidID)
	if err != nil {
		return domain.ProfileView{}, err
	}

	earned, err := s.composeEarned(ctx, q, kidID)
	if err != nil {
		return domain.ProfileView{}, err
	}

	next, err := s.composeNextAchievements(ctx, db, q, kidID)
	if err != nil {
		return domain.ProfileView{}, err
	}

	recent, err := s.composeRecentActivity(ctx, q, kidID)
	if err != nil {
		return domain.ProfileView{}, err
	}

	pending, err := s.composePendingRedemptions(ctx, q, kidID)
	if err != nil {
		return domain.ProfileView{}, err
	}

	return domain.ProfileView{
		Kid:                kid,
		Balance:            balance,
		CategoryLevels:     catLevels,
		NextAchievements:   next,
		RecentEarned:       earned,
		RecentActivity:     recent,
		PendingRedemptions: pending,
	}, nil
}

func (s *profileService) composeBalance(ctx context.Context, db storage.DBTX, kidID int64) (domain.Balance, error) {
	earned, err := s.balance.Earned(ctx, db, kidID)
	if err != nil {
		return domain.Balance{}, err
	}
	spent, err := s.balance.Spent(ctx, db, kidID)
	if err != nil {
		return domain.Balance{}, err
	}
	reserved, err := s.balance.Reserved(ctx, db, kidID)
	if err != nil {
		return domain.Balance{}, err
	}
	return domain.Balance{
		Earned:           earned,
		Spent:            spent,
		Reserved:         reserved,
		Balance:          earned - spent,
		AvailableBalance: earned - spent - reserved,
	}, nil
}

func (s *profileService) composeCategoryLevels(ctx context.Context, db storage.DBTX, q *sqldb.Queries, kidID int64) ([]domain.CategoryLevel, error) {
	cats, err := q.ListCategories(ctx)
	if err != nil {
		return nil, fmt.Errorf("list categories: %w", err)
	}
	out := make([]domain.CategoryLevel, 0, len(cats))
	for _, c := range cats {
		xp, err := q.SumActivityXPForKid(ctx, sqldb.SumActivityXPForKidParams{KidID: kidID, CategoryID: c.ID})
		if err != nil {
			return nil, fmt.Errorf("sum xp for category %d: %w", c.ID, err)
		}
		level := domain.LevelForXP(xp)
		floor := domain.XPForLevel(level)
		next := domain.XPForLevel(level + 1)
		ratio := 0.0
		if next > floor {
			ratio = float64(xp-floor) / float64(next-floor)
		}
		out = append(out, domain.CategoryLevel{
			Category: domain.Category{
				ID: c.ID, Slug: c.Slug, Name: c.Name,
				Description: c.Description, Icon: c.Icon, Color: c.Color,
				ArchivedAt: c.ArchivedAt, CreatedAt: c.CreatedAt,
			},
			XP:            xp,
			Level:         level,
			XPIntoLevel:   xp - floor,
			XPForNext:     next - floor,
			ProgressRatio: ratio,
		})
	}
	return out, nil
}

func (s *profileService) composeEarned(ctx context.Context, q *sqldb.Queries, kidID int64) ([]domain.EarnedAchievement, error) {
	rows, err := q.ListEarnedAchievementsForKid(ctx, kidID)
	if err != nil {
		return nil, fmt.Errorf("list earned: %w", err)
	}
	out := make([]domain.EarnedAchievement, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.EarnedAchievement{
			Achievement: domain.Achievement{
				ID: r.AchievementID, Slug: r.Slug, Name: r.Name,
				Description: r.Description, Title: r.Title,
				Combinator:  domain.Combinator(r.Combinator),
				BonusPoints: r.BonusPoints, ArchivedAt: r.ArchivedAt, CreatedAt: r.CreatedAt,
			},
			EarnedAt: r.EarnedAt,
			Unseen:   r.Unseen == 1,
		})
	}
	return out, nil
}

func (s *profileService) composeNextAchievements(ctx context.Context, db storage.DBTX, q *sqldb.Queries, kidID int64) ([]domain.AchievementProgress, error) {
	rows, err := q.ListUnearnedAchievementRulesForKid(ctx, kidID)
	if err != nil {
		return nil, fmt.Errorf("list unearned rules: %w", err)
	}
	achievements := GroupRulesByAchievement(rows)

	progressList := make([]domain.AchievementProgress, 0, len(achievements))
	for _, ach := range achievements {
		rules := make([]domain.RuleProgress, 0, len(ach.Rules))
		for _, rule := range ach.Rules {
			cur, err := ComputeRuleMetric(ctx, db, kidID, rule, s.balance)
			if err != nil {
				return nil, fmt.Errorf("compute metric for rule %d: %w", rule.ID, err)
			}
			ratio := 0.0
			if rule.Threshold > 0 {
				ratio = float64(cur) / float64(rule.Threshold)
				if ratio > 1 {
					ratio = 1
				}
			}
			rules = append(rules, domain.RuleProgress{
				Rule: rule, Current: cur, Threshold: rule.Threshold, Ratio: ratio,
			})
		}
		progressList = append(progressList, domain.AchievementProgress{
			Achievement:  ach,
			Rules:        rules,
			OverallRatio: CombinatorRatio(ach.Combinator, rules),
		})
	}

	sort.SliceStable(progressList, func(i, j int) bool {
		return progressList[i].OverallRatio > progressList[j].OverallRatio
	})
	if len(progressList) > nextAchievementLimit {
		progressList = progressList[:nextAchievementLimit]
	}
	return progressList, nil
}

func (s *profileService) composeRecentActivity(ctx context.Context, q *sqldb.Queries, kidID int64) ([]domain.Activity, error) {
	rows, err := q.ListRecentActivitiesForKid(ctx, sqldb.ListRecentActivitiesForKidParams{KidID: kidID, Limit: recentActivityLimit})
	if err != nil {
		return nil, fmt.Errorf("recent activity: %w", err)
	}
	out := make([]domain.Activity, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.Activity{
			ID: r.ID, KidID: r.KidID, ActivityTypeID: r.ActivityTypeID,
			Quantity: int(r.Quantity), XPAwarded: r.XpAwarded, PointsAwarded: r.PointsAwarded,
			Note: r.Note, OccurredAt: r.OccurredAt, CreatedAt: r.CreatedAt,
			VoidedAt: r.VoidedAt, VoidReason: r.VoidReason,
		})
	}
	return out, nil
}

func (s *profileService) composePendingRedemptions(ctx context.Context, q *sqldb.Queries, kidID int64) ([]domain.Redemption, error) {
	rows, err := q.ListPendingRedemptionsForKid(ctx, kidID)
	if err != nil {
		return nil, fmt.Errorf("pending redemptions: %w", err)
	}
	out := make([]domain.Redemption, 0, len(rows))
	for _, r := range rows {
		out = append(out, domain.Redemption{
			ID: r.ID, KidID: r.KidID, RewardID: r.RewardID,
			PointsSpent: r.PointsSpent, Status: domain.RedemptionStatus(r.Status),
			RequestedAt: r.RequestedAt, DecidedAt: r.DecidedAt,
		})
	}
	return out, nil
}

// Time-fenced helpers (placeholders for future window-scoped queries).
var _ = time.Now // keep time import for future helpers
