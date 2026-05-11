-- +goose Up
-- +goose StatementBegin

CREATE TABLE kids (
    id             INTEGER PRIMARY KEY AUTOINCREMENT,
    name           TEXT    NOT NULL,
    avatar_slug    TEXT    NOT NULL,
    color          TEXT    NOT NULL DEFAULT '#6366F1',
    display_order  INTEGER NOT NULL DEFAULT 0,
    archived_at    TIMESTAMP,
    created_at     TIMESTAMP NOT NULL DEFAULT (datetime('now'))
);
CREATE INDEX idx_kids_display_order ON kids(display_order) WHERE archived_at IS NULL;

CREATE TABLE categories (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    slug         TEXT    NOT NULL,
    name         TEXT    NOT NULL,
    description  TEXT,
    icon         TEXT,
    color        TEXT,
    archived_at  TIMESTAMP,
    created_at   TIMESTAMP NOT NULL DEFAULT (datetime('now'))
);
CREATE UNIQUE INDEX idx_categories_slug_active ON categories(slug) WHERE archived_at IS NULL;

CREATE TABLE activity_types (
    id               INTEGER PRIMARY KEY AUTOINCREMENT,
    category_id      INTEGER NOT NULL REFERENCES categories(id),
    slug             TEXT    NOT NULL,
    name             TEXT    NOT NULL,
    description      TEXT,
    xp_per_unit      INTEGER NOT NULL CHECK (xp_per_unit >= 0),
    points_per_unit  INTEGER NOT NULL CHECK (points_per_unit >= 0),
    archived_at      TIMESTAMP,
    created_at       TIMESTAMP NOT NULL DEFAULT (datetime('now'))
);
CREATE UNIQUE INDEX idx_activity_types_slug_active ON activity_types(slug) WHERE archived_at IS NULL;
CREATE INDEX idx_activity_types_category ON activity_types(category_id) WHERE archived_at IS NULL;

CREATE TABLE achievements (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    slug          TEXT    NOT NULL,
    name          TEXT    NOT NULL,
    description   TEXT,
    title         TEXT,
    combinator    TEXT    NOT NULL CHECK (combinator IN ('ALL', 'ANY')),
    bonus_points  INTEGER NOT NULL DEFAULT 0 CHECK (bonus_points >= 0),
    archived_at   TIMESTAMP,
    created_at    TIMESTAMP NOT NULL DEFAULT (datetime('now'))
);
CREATE UNIQUE INDEX idx_achievements_slug_active ON achievements(slug) WHERE archived_at IS NULL;

CREATE TABLE achievement_rules (
    id              INTEGER PRIMARY KEY AUTOINCREMENT,
    achievement_id  INTEGER NOT NULL REFERENCES achievements(id) ON DELETE CASCADE,
    category_id     INTEGER          REFERENCES categories(id),
    metric          TEXT    NOT NULL CHECK (metric IN ('count', 'xp', 'points', 'level')),
    threshold       INTEGER NOT NULL CHECK (threshold > 0)
);
CREATE INDEX idx_achievement_rules_achievement ON achievement_rules(achievement_id);

CREATE TABLE rewards (
    id           INTEGER PRIMARY KEY AUTOINCREMENT,
    slug         TEXT    NOT NULL,
    name         TEXT    NOT NULL,
    description  TEXT,
    cost_points  INTEGER NOT NULL CHECK (cost_points > 0),
    active       INTEGER NOT NULL DEFAULT 1 CHECK (active IN (0, 1)),
    archived_at  TIMESTAMP,
    created_at   TIMESTAMP NOT NULL DEFAULT (datetime('now'))
);
CREATE UNIQUE INDEX idx_rewards_slug_active ON rewards(slug) WHERE archived_at IS NULL;

CREATE TABLE activities (
    id                INTEGER PRIMARY KEY AUTOINCREMENT,
    kid_id            INTEGER NOT NULL REFERENCES kids(id),
    activity_type_id  INTEGER NOT NULL REFERENCES activity_types(id),
    quantity          INTEGER NOT NULL DEFAULT 1 CHECK (quantity > 0),
    xp_awarded        INTEGER NOT NULL CHECK (xp_awarded >= 0),
    points_awarded    INTEGER NOT NULL CHECK (points_awarded >= 0),
    note              TEXT,
    occurred_at       TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    created_at        TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    voided_at         TIMESTAMP,
    void_reason       TEXT
);
CREATE INDEX idx_activities_kid_occurred ON activities(kid_id, occurred_at);
CREATE INDEX idx_activities_kid_type_active ON activities(kid_id, activity_type_id) WHERE voided_at IS NULL;

CREATE TABLE kid_achievements (
    kid_id          INTEGER NOT NULL REFERENCES kids(id),
    achievement_id  INTEGER NOT NULL REFERENCES achievements(id),
    earned_at       TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    unseen          INTEGER NOT NULL DEFAULT 1 CHECK (unseen IN (0, 1)),
    PRIMARY KEY (kid_id, achievement_id)
);
CREATE INDEX idx_kid_achievements_unseen ON kid_achievements(kid_id, unseen) WHERE unseen = 1;

CREATE TABLE redemptions (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    kid_id        INTEGER NOT NULL REFERENCES kids(id),
    reward_id     INTEGER NOT NULL REFERENCES rewards(id),
    points_spent  INTEGER NOT NULL CHECK (points_spent > 0),
    status        TEXT    NOT NULL CHECK (status IN ('pending', 'approved', 'rejected', 'cancelled')),
    requested_at  TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    decided_at    TIMESTAMP
);
CREATE INDEX idx_redemptions_kid_status ON redemptions(kid_id, status);

CREATE TABLE point_adjustments (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    kid_id        INTEGER NOT NULL REFERENCES kids(id),
    points_delta  INTEGER NOT NULL CHECK (points_delta != 0),
    reason        TEXT,
    created_at    TIMESTAMP NOT NULL DEFAULT (datetime('now')),
    voided_at     TIMESTAMP,
    void_reason   TEXT
);
CREATE INDEX idx_point_adjustments_kid_active ON point_adjustments(kid_id) WHERE voided_at IS NULL;

-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS point_adjustments;
DROP TABLE IF EXISTS redemptions;
DROP TABLE IF EXISTS kid_achievements;
DROP TABLE IF EXISTS activities;
DROP TABLE IF EXISTS rewards;
DROP TABLE IF EXISTS achievement_rules;
DROP TABLE IF EXISTS achievements;
DROP TABLE IF EXISTS activity_types;
DROP TABLE IF EXISTS categories;
DROP TABLE IF EXISTS kids;
-- +goose StatementEnd
