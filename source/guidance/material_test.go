package guidance

import (
	"reflect"
	"strings"
	"testing"
)

func TestMaterialComponentPages(t *testing.T) {
	source := []string{"https://m3.material.io/components/buttons/specs", "https://m3.material.io/components/buttons/specs", "https://m3.material.io/styles/color/tokens", "https://m3.material.io/components/navigation-rail/specs"}
	pages, err := MaterialComponentPages(source, []string{"navigation-rail"})
	want := []string{"https://m3.material.io/components/navigation-rail/overview", "https://m3.material.io/components/navigation-rail/specs", "https://m3.material.io/components/navigation-rail/guidelines", "https://m3.material.io/components/navigation-rail/accessibility"}
	if err != nil || !reflect.DeepEqual(pages, want) {
		t.Fatalf("got %v %v", pages, err)
	}
	pages, err = MaterialComponentPages(source, nil)
	if err != nil || len(pages) != 8 {
		t.Fatalf("all components: %v %v", pages, err)
	}
	if _, err = MaterialComponentPages(source, []string{"unknown"}); err == nil || !strings.Contains(err.Error(), "absent") {
		t.Fatalf("unknown component: %v", err)
	}
	if _, err = MaterialComponentPages(source[2:3], nil); err == nil {
		t.Fatal("accepted catalog with no components")
	}
}
