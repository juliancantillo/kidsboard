/** @type {import('tailwindcss').Config} */
// Mirrors the inline `tailwind.config = {...}` block that used to live in
// internal/view/templates/layouts/base.html. Standalone Tailwind v3 CLI
// (no Node required) consumes this file when invoked via `make css` and
// the Dockerfile build stage. The forms + container-queries plugins are
// bundled into the standalone binary so they're available without an
// explicit `plugins:` entry here.
module.exports = {
  content: ["./internal/view/templates/**/*.html"],
  theme: {
    extend: {
      colors: {
        "ink":            "#261900",
        "ink-soft":       "#3f2e06",
        "parchment":      "#fff8f2",
        "surface":        "#fff8f2",
        "surface-dim":    "#f4d6a0",
        "surface-low":    "#fff2df",
        "surface-mid":    "#ffebcc",
        "surface-high":   "#ffe5b8",
        "surface-edge":   "#fddfa8",
        "wood":           "#7d5233",
        "wood-dark":      "#623b1e",
        "wood-shadow":    "#301400",
        "wood-light":     "#a47349",
        "ember":          "#e07a3f",
        "ember-soft":     "#ffcaa7",
        "moss":           "#0f6c43",
        "moss-soft":      "#9df2be",
        "moss-bright":    "#1ea968",
        "river":          "#004a61",
        "river-soft":     "#9dddfd",
        "blood":          "#ba1a1a",
        "blood-soft":     "#ffdad6",
        "gold":           "#d9a441",
        "outline":        "#83746b",
        "outline-soft":   "#d5c3b9",
      },
      fontFamily: {
        display: ["Anybody", "system-ui", "sans-serif"],
        body:    ["Courier Prime", "ui-monospace", "monospace"],
        label:   ["JetBrains Mono", "ui-monospace", "monospace"],
      },
      boxShadow: {
        "pixel-sm":    "3px 3px 0 0 #261900",
        "pixel":       "4px 4px 0 0 #261900",
        "pixel-lg":    "6px 6px 0 0 #261900",
        "pixel-xl":    "8px 8px 0 0 #261900",
        "pixel-inset": "inset 0 4px 0 0 rgba(255,255,255,.30), inset 0 -3px 0 0 rgba(0,0,0,.30)",
      },
    },
  },
};
