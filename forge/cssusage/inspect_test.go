package cssusage

import (
	"strings"
	"testing"
)

func TestVariableChainsAndManualValues(t *testing.T) {
	report := Inspect(map[string]string{"menu.css": "/* upstream */\n.md-menu {\n padding:var(--_padding);\n display:flex;\n gap:var(--unknown);\n color:var(--_color);\n}"}, ".md-menu {--_padding:var(--md-comp-menus-group-padding);--_color:transparent;} .md-menu:hover {--_color:var(--md-sys-color-primary);}", ":root{--md-comp-menus-group-padding:2px;--md-sys-color-primary:#fff;} :root[data-theme=dark]{--md-comp-menus-group-padding:2px;--md-sys-color-primary:#000;}")
	if report.Count("Token") != 1 || report.Count("Mixed") != 1 || report.Count("Manual") != 1 || report.Count("Unresolved") != 1 {
		t.Fatalf("incorrect classification: %#v", report)
	}
	if report.Declarations[0].Line != 3 || report.Declarations[0].Tokens[0] != "--md-comp-menus-group-padding" {
		t.Fatal("lost source line or variable chain", report.Declarations[0])
	}
	mixed := report.Declarations[3]
	if len(mixed.ManualValues) != 1 || mixed.ManualValues[0].Value != "transparent" || mixed.ManualValues[0].Selector != ".md-menu" || mixed.ManualValues[0].Property != "--_color" {
		t.Fatalf("missing manual definition: %#v", mixed)
	}
	if got := report.Declarations[0].TokenValues; len(got) != 1 || got[0].Value != "2px" || !strings.Contains(got[0].Selector, "dark") {
		t.Fatalf("missing source value: %#v", got)
	}
	if len(report.ForFamily("menu").Declarations) != 4 || len(report.ForFamily("button").Declarations) != 0 {
		t.Fatal("incorrect family filter")
	}
}
func TestNestedRulesStringsFallbacksAndCycles(t *testing.T) {
	report := Inspect(map[string]string{"test.css": "@media (prefers-color-scheme: dark) {.md-menu {content:\"x;y\"; padding:var(--md-sys-space, 4px);color:var(--cycle);}}"}, ".x{--cycle:var(--cycle);}", ":root{--md-sys-space:2px;}")
	if len(report.Declarations) != 3 || report.Count("Mixed") != 1 || report.Count("Unresolved") != 1 {
		t.Fatal(report)
	}
	if got := report.Declarations[1].ManualValues; len(got) != 1 || got[0].Value != "4px" || got[0].Reason != "fallback" || got[0].File != "test.css" {
		t.Fatalf("missing fallback definition: %#v", got)
	}
	if report.Declarations[0].Value != "\"x;y\"" || report.Declarations[0].Selector != ".md-menu" {
		t.Fatal(report.Declarations[0])
	}
}
