package main

import (
	"strings"
	"testing"
)

func TestCleanImportsPreservesUsedAndSideEffectImports(t *testing.T) {
	input := []byte(`package render
import "fmt"
import "fmt"
import "strings"
import _ "net/http/pprof"
func render() string { return fmt.Sprint("strings is text, not an import use") }
`)
	output, err := cleanImports(input)
	if err != nil {
		t.Fatal(err)
	}
	text := string(output)
	if strings.Count(text, `"fmt"`) != 1 || strings.Contains(text, `import "strings"`) || !strings.Contains(text, `_ "net/http/pprof"`) {
		t.Fatalf("unexpected imports:\n%s", text)
	}
}
