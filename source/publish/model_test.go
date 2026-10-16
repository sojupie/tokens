package publish

import "testing"

func TestSubjectKeyStringIncludesKind(t *testing.T) {
	component := SubjectKey{Kind: SubjectComponent, Slug: "elevation"}
	style := SubjectKey{Kind: SubjectStyle, Slug: "elevation"}
	if component.String() == style.String() {
		t.Fatalf("subject keys must distinguish kinds: %q", component.String())
	}
	if got, want := component.String(), "component:elevation"; got != want {
		t.Fatalf("component key = %q, want %q", got, want)
	}
}
