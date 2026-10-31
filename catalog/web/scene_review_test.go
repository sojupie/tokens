package web

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestSceneReviewKeepsNativeViewportAndImagesSeparate(t *testing.T) {
	server := New(seoTestState(), "")
	server.thumbnails = map[string]componentCase{"snackbar": {Family: "snackbar", Variant: "two-line", State: "enabled"}}
	page := httptest.NewRecorder()
	server.Handler().ServeHTTP(page, httptest.NewRequest("GET", "/illustrations/review/components", nil))
	if page.Code != 200 {
		t.Fatal(page.Code, page.Body.String())
	}
	for _, want := range []string{`name="robots" content="noindex"`, `data-review-grid`, `/illustrations/render/components/snackbar?view=review`, `width="412" height="892"`, `loading="lazy"`} {
		if !strings.Contains(page.Body.String(), want) {
			t.Errorf("review missing %s", want)
		}
	}
	scene := httptest.NewRecorder()
	server.Handler().ServeHTTP(scene, httptest.NewRequest("GET", "/illustrations/render/components/snackbar?view=review", nil))
	for _, want := range []string{`data-scene-review`, `scene-column-overlay tokens-layout-content`, `data-device="phone"`, `Conversation archived`, `<h1 class="md-appbar-title">Inbox</h1>`} {
		if !strings.Contains(scene.Body.String(), want) {
			t.Errorf("scene missing %s", want)
		}
	}
	if strings.Contains(scene.Body.String(), "illustration-ready.js") {
		t.Error("review must not freeze motion or start image capture readiness")
	}
	if strings.Contains(scene.Body.String(), `class="scene-heading`) {
		t.Error("generic duplicate scene heading returned")
	}
}

func TestSceneStateMatchesVisibleTask(t *testing.T) {
	if len(sceneInboxRows(false)) != 8 || len(sceneInboxRows(true)) != 7 || sceneInboxRows(true)[0][1] == sceneInboxRows(false)[0][1] {
		t.Fatal("archive scene must actually remove a conversation")
	}
	calendar := scenePreset(componentCase{Family: "segmented"})
	if !calendar.ItemSelected(1) || calendar.GroupLabel(1) != "Week" {
		t.Fatal("weekly calendar must select Week")
	}
	orders := scenePreset(componentCase{Family: "tabs"})
	if !orders.ItemSelected(0) || orders.GroupLabel(0) != "Active" {
		t.Fatal("active orders must select Active")
	}
	delivery := scenePreset(componentCase{Family: "radio"})
	if delivery.RadioInput(0)["checked"] != true {
		t.Fatal("total includes standard delivery, which must be selected")
	}
}

func TestSceneHeaderIsChosenForTheTask(t *testing.T) {
	for _, family := range []string{"badge", "menu", "tooltip", "toolbar", "card", "chip", "segmented", "nav-drawer", "progress", "field", "split"} {
		if sceneUsesAppBar(family) {
			t.Errorf("%s must not inherit an app bar", family)
		}
	}
	if !sceneUsesAppBar("appbar") {
		t.Fatal("the app bar scene needs its actual primitive")
	}
}
