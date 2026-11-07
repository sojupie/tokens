package components

import (
	"encoding/json"
	"fmt"
	"sync"
)

var scriptModules = sync.OnceValues(func() (map[string]struct{ Scripts []string }, error) {
	data, err := Files.ReadFile("modules.json")
	if err != nil {
		return nil, err
	}
	var modules map[string]struct{ Scripts []string }
	err = json.Unmarshal(data, &modules)
	return modules, err
})

// ScriptsFor resolves ownership in the caller's module order. It does not infer
// execution prerequisites or activation policy from component ownership.
func ScriptsFor(names ...string) ([]string, error) {
	modules, err := scriptModules()
	if err != nil {
		return nil, err
	}
	var scripts []string
	seen := make(map[string]bool)
	for _, name := range names {
		module, ok := modules[name]
		if !ok {
			return nil, fmt.Errorf("unknown component module: %s", name)
		}
		for _, script := range module.Scripts {
			if !seen[script] {
				scripts = append(scripts, script)
				seen[script] = true
			}
		}
	}
	return scripts, nil
}
