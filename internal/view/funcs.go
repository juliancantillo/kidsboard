package view

import (
	"fmt"
	"html/template"
	"strings"
	"time"
)

// funcMap registers helpers callable from any template.
func funcMap() template.FuncMap {
	return template.FuncMap{
		"avatarSrc":    avatarSrc,
		"bgSrc":        bgSrc,
		"iconSrc":      iconSrc,
		"upper":        strings.ToUpper,
		"pct":          pct,
		"clampPct":     clampPct,
		"int":          toInt,
		"fmtInt":       fmtInt,
		"fmtTime":      fmtTime,
		"fmtRelative":  fmtRelative,
		"deref":        deref,
		"defaultIfNil": defaultIfNil,
		"add":          func(a, b int) int { return a + b },
		"dict":         dict,
	}
}

func avatarSrc(slug string) string {
	if slug == "" {
		return "/static/avatars/red-fairy.png"
	}
	return "/static/avatars/" + slug + ".png"
}

func bgSrc(name string) string {
	return "/static/bg/" + name + ".png"
}

// iconSrc resolves a category/UI icon slug to its static path. Slugs match
// files under static/icons/ exactly (without extension). Empty slug returns
// an empty string so templates can branch on `{{ with iconSrc .Icon }}`.
func iconSrc(slug string) string {
	if slug == "" {
		return ""
	}
	return "/static/icons/" + slug + ".png"
}

func pct(value float64) string {
	return fmt.Sprintf("%.0f%%", value*100)
}

// clampPct keeps a 0..1 ratio safe for width: X% styles.
func clampPct(value float64) string {
	if value < 0 {
		value = 0
	}
	if value > 1 {
		value = 1
	}
	return fmt.Sprintf("%.0f%%", value*100)
}

func toInt(v int64) int { return int(v) }

func fmtInt(v int64) string {
	// Thin space separator for readability: 12 200 instead of 12,200.
	s := fmt.Sprintf("%d", v)
	if len(s) <= 3 {
		return s
	}
	// Insert thin spaces every 3 digits from the right.
	var b strings.Builder
	pre := len(s) % 3
	if pre > 0 {
		b.WriteString(s[:pre])
		if len(s) > pre {
			b.WriteString(" ")
		}
	}
	for i := pre; i < len(s); i += 3 {
		b.WriteString(s[i : i+3])
		if i+3 < len(s) {
			b.WriteString(" ")
		}
	}
	return b.String()
}

func fmtTime(t time.Time) string {
	return t.Format("2 Jan, 15:04")
}

func fmtRelative(t time.Time) string {
	d := time.Since(t)
	switch {
	case d < time.Minute:
		return "ahora"
	case d < time.Hour:
		return fmt.Sprintf("hace %d min", int(d.Minutes()))
	case d < 24*time.Hour:
		return fmt.Sprintf("hace %d h", int(d.Hours()))
	case d < 7*24*time.Hour:
		return fmt.Sprintf("hace %d d", int(d.Hours()/24))
	default:
		return t.Format("2 Jan")
	}
}

func deref(s *string) string {
	if s == nil {
		return ""
	}
	return *s
}

func defaultIfNil(s *string, fallback string) string {
	if s == nil || *s == "" {
		return fallback
	}
	return *s
}

func dict(values ...any) (map[string]any, error) {
	if len(values)%2 != 0 {
		return nil, fmt.Errorf("dict requires an even number of arguments")
	}
	m := make(map[string]any, len(values)/2)
	for i := 0; i < len(values); i += 2 {
		key, ok := values[i].(string)
		if !ok {
			return nil, fmt.Errorf("dict keys must be strings")
		}
		m[key] = values[i+1]
	}
	return m, nil
}
