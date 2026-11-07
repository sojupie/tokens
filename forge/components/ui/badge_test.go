package ui

import "testing"

func TestBadgeContent(t *testing.T) {
	for input, want := range map[string]string{"1": "1", "12": "12", "999": "999", "1000": "999+", "99999": "999+", "18446744073709551616": "999+", "99+": "99+", "New": "New", "Ready": "Read", " 新着 ": "新着"} {
		if got := BadgeContent(input); got != want {
			t.Errorf("BadgeContent(%q)=%q, want %q", input, got, want)
		}
	}
}
