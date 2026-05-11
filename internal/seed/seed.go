package seed

import (
	"context"
	"database/sql"
	"fmt"

	"cantillo.dev/kidsboard/internal/storage"
	"cantillo.dev/kidsboard/internal/storage/sqldb"
)

// Run upserts the curated seed data into the DB. Idempotent — re-running
// updates fields by slug without touching IDs or event-table FKs.
func Run(ctx context.Context, db *sql.DB) error {
	_, err := storage.WithTx(ctx, db, func(tx storage.DBTX) (struct{}, error) {
		q := sqldb.New(tx)

		catBySlug, err := upsertCategories(ctx, q)
		if err != nil {
			return struct{}{}, err
		}
		if err := upsertActivityTypes(ctx, q, catBySlug); err != nil {
			return struct{}{}, err
		}
		if err := upsertAchievements(ctx, q, catBySlug); err != nil {
			return struct{}{}, err
		}
		return struct{}{}, nil
	})
	return err
}

func upsertCategories(ctx context.Context, q *sqldb.Queries) (map[string]int64, error) {
	out := make(map[string]int64, len(Categories))
	for _, c := range Categories {
		desc := optionalString(c.Description)
		icon := optionalString(c.Icon)
		color := optionalString(c.Color)
		row, err := q.UpsertCategoryBySlug(ctx, sqldb.UpsertCategoryBySlugParams{
			Slug: c.Slug, Name: c.Name,
			Description: desc, Icon: icon, Color: color,
		})
		if err != nil {
			return nil, fmt.Errorf("upsert category %q: %w", c.Slug, err)
		}
		out[c.Slug] = row.ID
	}
	return out, nil
}

func upsertActivityTypes(ctx context.Context, q *sqldb.Queries, catBySlug map[string]int64) error {
	for _, a := range ActivityTypes {
		catID, ok := catBySlug[a.CategorySlug]
		if !ok {
			return fmt.Errorf("activity_type %q references unknown category %q", a.Slug, a.CategorySlug)
		}
		desc := optionalString(a.Description)
		if _, err := q.UpsertActivityTypeBySlug(ctx, sqldb.UpsertActivityTypeBySlugParams{
			CategoryID: catID, Slug: a.Slug, Name: a.Name,
			Description:   desc,
			XpPerUnit:     a.XPPerUnit,
			PointsPerUnit: a.PointsPerUnit,
		}); err != nil {
			return fmt.Errorf("upsert activity_type %q: %w", a.Slug, err)
		}
	}
	return nil
}

func upsertAchievements(ctx context.Context, q *sqldb.Queries, catBySlug map[string]int64) error {
	for _, ach := range Achievements {
		desc := optionalString(ach.Description)
		title := optionalString(ach.Title)
		row, err := q.UpsertAchievementBySlug(ctx, sqldb.UpsertAchievementBySlugParams{
			Slug: ach.Slug, Name: ach.Name,
			Description: desc, Title: title,
			Combinator:  ach.Combinator,
			BonusPoints: ach.BonusPoints,
		})
		if err != nil {
			return fmt.Errorf("upsert achievement %q: %w", ach.Slug, err)
		}
		// Reset rules: simpler than diffing. Achievements are seeded, so the
		// in-code spec is authoritative. Event tables (kid_achievements) are
		// unaffected because they reference the achievement, not its rules.
		if err := q.DeleteAchievementRules(ctx, row.ID); err != nil {
			return fmt.Errorf("clear rules for %q: %w", ach.Slug, err)
		}
		for _, r := range ach.Rules {
			var catID *int64
			if r.CategorySlug != "" {
				id, ok := catBySlug[r.CategorySlug]
				if !ok {
					return fmt.Errorf("achievement %q rule references unknown category %q", ach.Slug, r.CategorySlug)
				}
				catID = &id
			}
			if _, err := q.InsertAchievementRule(ctx, sqldb.InsertAchievementRuleParams{
				AchievementID: row.ID,
				CategoryID:    catID,
				Metric:        r.Metric,
				Threshold:     r.Threshold,
			}); err != nil {
				return fmt.Errorf("insert rule for %q: %w", ach.Slug, err)
			}
		}
	}
	return nil
}

func optionalString(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}

// Stats summarizes what's in the database — used by the CLI to confirm seeding.
type Stats struct {
	Categories       int
	ActivityTypes    int
	Achievements     int
	AchievementRules int
}

func Count(ctx context.Context, db *sql.DB) (Stats, error) {
	q := sqldb.New(db)
	cats, err := q.ListCategories(ctx)
	if err != nil {
		return Stats{}, err
	}
	types, err := q.ListActivityTypes(ctx)
	if err != nil {
		return Stats{}, err
	}
	achs, err := q.ListAchievements(ctx)
	if err != nil {
		return Stats{}, err
	}
	ruleCount := 0
	for _, a := range achs {
		rs, err := q.ListAchievementRules(ctx, a.ID)
		if err != nil {
			return Stats{}, err
		}
		ruleCount += len(rs)
	}
	return Stats{
		Categories:       len(cats),
		ActivityTypes:    len(types),
		Achievements:     len(achs),
		AchievementRules: ruleCount,
	}, nil
}
