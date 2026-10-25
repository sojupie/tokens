package guidance

import (
	"fmt"
	"net/url"
	"strings"
)

// MaterialComponentPages expands the token crawler's component specs URLs into
// Material's fixed documentation tabs. Style/foundation pages are not components.
func MaterialComponentPages(tokenPages, components []string) ([]string, error) {
	wanted := map[string]bool{}
	for _, component := range components {
		wanted[component] = false
	}
	seen := map[string]bool{}
	var pages []string
	for _, raw := range tokenPages {
		u, err := url.Parse(raw)
		if err != nil {
			return nil, err
		}
		if !strings.HasPrefix(u.Path, "/components/") || !strings.HasSuffix(u.Path, "/specs") {
			continue
		}
		component := strings.TrimSuffix(strings.TrimPrefix(u.Path, "/components/"), "/specs")
		if len(wanted) > 0 {
			if _, ok := wanted[component]; !ok {
				continue
			}
			wanted[component] = true
		}
		root := strings.TrimSuffix(u.Path, "/specs")
		u.RawQuery = ""
		u.Fragment = ""
		u.RawPath = ""
		for _, tab := range []string{"overview", "specs", "guidelines", "accessibility"} {
			u.Path = root + "/" + tab
			page := u.String()
			if !seen[page] {
				seen[page] = true
				pages = append(pages, page)
			}
		}
	}
	for _, component := range components {
		if !wanted[component] {
			return nil, fmt.Errorf("component %q is absent from the page catalog", component)
		}
	}
	if len(pages) == 0 {
		return nil, fmt.Errorf("page catalog contains no selected component specs pages")
	}
	return pages, nil
}
