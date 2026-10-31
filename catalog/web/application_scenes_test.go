package web

import (
	"golang.org/x/net/html"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestApplicationScenesKeepCoverageAndNativeSizes(t *testing.T) {
	server := New(seoTestState(), "")
	handler := server.Handler()
	if len(applicationScenes) != 6 {
		t.Fatal("expected six applications")
	}
	views := 0
	for _, app := range applicationScenes {
		for _, view := range app.Views {
			views++
			t.Run(app.ID+"/"+view.ID, func(t *testing.T) {
				response := httptest.NewRecorder()
				handler.ServeHTTP(response, httptest.NewRequest("GET", app.URL(view.ID), nil))
				if response.Code != 200 {
					t.Fatal(response.Code, response.Body.String())
				}
				body := response.Body.String()
				for _, want := range []string{`data-application="` + app.ID + `"`, `data-view="` + view.ID + `"`, `name="robots" content="noindex"`, `application-scenes.css`} {
					if !strings.Contains(body, want) {
						t.Errorf("missing %s", want)
					}
				}
				if strings.Contains(body, "illustration-ready.js") {
					t.Error("live review must not freeze or capture images")
				}
				if app.ID == "museum" && !strings.Contains(body, "--app-width:412px;--app-height:892px") {
					t.Error("museum must retain phone dimensions")
				}
			})
		}
	}
	if views != 22 {
		t.Fatalf("got %d states", views)
	}
	for _, path := range []string{"/illustrations/apps/unknown/inbox", "/illustrations/apps/mail/missing"} {
		response := httptest.NewRecorder()
		handler.ServeHTTP(response, httptest.NewRequest("GET", path, nil))
		if response.Code != 404 {
			t.Errorf("%s: %d", path, response.Code)
		}
	}
	review := httptest.NewRecorder()
	handler.ServeHTTP(review, httptest.NewRequest("GET", "/illustrations/review", nil))
	if strings.Count(review.Body.String(), "<iframe ") != 6 {
		t.Error("review needs one frame per application")
	}
	if strings.Count(review.Body.String(), "data-app-view=") != 22 {
		t.Error("review must expose every state")
	}
}

func TestApplicationSceneContracts(t *testing.T) {
	handler := New(seoTestState(), "").Handler()
	render := func(path string) string {
		t.Helper()
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 200 {
			t.Fatalf("%s: %d", path, w.Code)
		}
		return w.Body.String()
	}
	mail := render("/illustrations/apps/mail/inbox")
	if !strings.Contains(mail, `md-navigation-rail`) || strings.Contains(mail, `md-navigation-drawer`) {
		t.Error("Compose belongs to Mail's rail, not a drawer")
	}
	workspace := render("/illustrations/apps/workspace/tasks")
	if !strings.Contains(workspace, `md-navigation-drawer`) || strings.Contains(workspace, `md-nav-fab`) {
		t.Error("workspace drawer must contain navigation without a FAB")
	}
	for _, app := range applicationScenes {
		for _, view := range app.Views {
			body := render(app.URL(view.ID))
			for _, bad := range []string{`app-modal-layer`, `scene-modal-shade`, `component-preview`, `illustration-ready.js`} {
				if strings.Contains(body, bad) {
					t.Errorf("%s/%s contains playground/capture scaffold %s", app.ID, view.ID, bad)
				}
			}
			if !strings.Contains(body, `asset-loader.js`) {
				t.Errorf("%s/%s bypasses shared component delivery", app.ID, view.ID)
			}
		}
	}
	dialog := render("/illustrations/apps/workspace/delete")
	if !strings.Contains(dialog, `data-dialog-surface`) || strings.Contains(dialog, `data-dialog-preview`) {
		t.Error("native dialog must retain surface sizing without the edit-demo controller")
	}
	details := render("/illustrations/apps/artwork/details?item=2")
	if !strings.Contains(details, `Lake Lucerne</h2>`) || !strings.Contains(details, `md-sheet-host`) {
		t.Error("details must retain the chosen artwork and docked sheet host")
	}
	checkout := render("/illustrations/apps/shop/checkout?size=2&frame=2")
	for _, want := range []string{`50 × 70 cm · Oak`, `Quantity 1 · €86.00`, `€90.90`} {
		if !strings.Contains(checkout, want) {
			t.Errorf("checkout lost chosen options: %s", want)
		}
	}
	for _, path := range []string{"/illustrations/apps/mail/inbox?item=-1", "/illustrations/apps/artwork/details?item=9", "/illustrations/apps/mail/inbox?width=1", "/illustrations/apps/shop/checkout?frame=7"} {
		w := httptest.NewRecorder()
		handler.ServeHTTP(w, httptest.NewRequest("GET", path, nil))
		if w.Code != 400 {
			t.Errorf("%s: got %d, want 400", path, w.Code)
		}
	}
}

func TestApplicationSceneAccessibleReferences(t *testing.T) {
	handler := New(seoTestState(), "").Handler()
	for _, app := range applicationScenes {
		for _, view := range app.Views {
			t.Run(app.ID+"/"+view.ID, func(t *testing.T) {
				w := httptest.NewRecorder()
				handler.ServeHTTP(w, httptest.NewRequest("GET", app.URL(view.ID), nil))
				doc, err := html.ParseWithOptions(strings.NewReader(w.Body.String()), html.ParseOptionEnableScripting(false))
				if err != nil {
					t.Fatal(err)
				}
				ids := map[string]bool{}
				var refs []string
				var visit func(*html.Node)
				visit = func(n *html.Node) {
					attrs := map[string]bool{}
					for _, a := range n.Attr {
						if attrs[a.Key] {
							t.Errorf("duplicate %s attribute on %s", a.Key, n.Data)
						}
						attrs[a.Key] = true
						if a.Key == "id" {
							if ids[a.Val] {
								t.Errorf("duplicate id %s", a.Val)
							}
							ids[a.Val] = true
						}
						if a.Key == "aria-controls" || a.Key == "aria-labelledby" || a.Key == "aria-describedby" || a.Key == "for" {
							refs = append(refs, strings.Fields(a.Val)...)
						}
					}
					for c := n.FirstChild; c != nil; c = c.NextSibling {
						visit(c)
					}
				}
				visit(doc)
				for _, id := range refs {
					if !ids[id] {
						t.Errorf("missing accessible reference %s", id)
					}
				}
			})
		}
	}
}
