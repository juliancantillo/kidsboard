package view

import "embed"

// Static is the embedded FS for everything under /static — avatars,
// backgrounds, generated CSS. Mounted by the HTTP server at /static/.
//
//go:embed static
var Static embed.FS

// Avatars is the curated list of pixel-art avatar slugs available to kids.
// Slugs must match a file under static/avatars/<slug>.png exactly. Edit this
// list (and add the PNG) to expand the roster.
var Avatars = []Avatar{
	{Slug: "red-fairy", Name: "Hada Carmesí"},
	{Slug: "blue-fairy", Name: "Hada Zafiro"},
	{Slug: "ember-warrior", Name: "Guerrera de Ascuas"},
	{Slug: "green-druid", Name: "Druida del Bosque"},
	{Slug: "violet-scholar", Name: "Erudito Violeta"},
	{Slug: "golden-mage", Name: "Maga Dorada"},
	{Slug: "sunlit-rogue", Name: "Pícaro del Sol"},
	{Slug: "teal-ranger", Name: "Explorador Turquesa"},
	{Slug: "silver-knight", Name: "Caballero de Plata"},
	{Slug: "copper-bard", Name: "Bardo de Cobre"},
	{Slug: "midnight-mystic", Name: "Mística Nocturna"},
	{Slug: "storm-archer", Name: "Arquero de la Tormenta"},
	{Slug: "sage-healer", Name: "Sanadora Sabia"},
	{Slug: "crimson-paladin", Name: "Paladina Carmesí"},
}

// Avatar describes one entry in the curated roster.
type Avatar struct {
	Slug string
	Name string
}

// AvatarSlugs returns the slug set as a fast-lookup map for validators.
func AvatarSlugs() map[string]struct{} {
	m := make(map[string]struct{}, len(Avatars))
	for _, a := range Avatars {
		m[a.Slug] = struct{}{}
	}
	return m
}
