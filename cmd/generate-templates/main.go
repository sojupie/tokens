// generate-templates compiles Catalog and Forge templates into generated-only render packages.
package main

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/format"
	"go/parser"
	"go/token"
	"io/fs"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

func main() {
	check := flag.Bool("check", false, "verify generated files without changing them")
	flag.Parse()
	if err := generate(*check); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func generate(check bool) error {
	if err := generateAt("catalog/web/templates", "catalog/web/internal/render", check); err != nil {
		return err
	}
	if err := generateAt("forge/components/ui/templates", "forge/components/ui/internal/render", check); err != nil {
		return err
	}
	return generateUIFacade(check)
}

func generateAt(source, output string, check bool) error {
	root, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		if _, err := os.Stat(filepath.Join(root, "go.mod")); err == nil {
			break
		}
		parent := filepath.Dir(root)
		if parent == root {
			return fmt.Errorf("run inside the Token Catalog repository")
		}
		root = parent
	}
	if err := os.Chdir(root); err != nil {
		return err
	}
	// Resolve the pinned tool once, rather than compiling or locating it per file.
	command := exec.Command("go", "tool", "-n", "templ")
	tool, err := command.Output()
	if err != nil {
		return fmt.Errorf("resolve templ: %w", err)
	}

	expected := map[string]bool{}
	err = filepath.WalkDir(source, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !entry.IsDir() && strings.HasSuffix(path, "_templ.go") {
			return fmt.Errorf("%s: generated Go belongs in internal/render; remove this file and run npm run generate:web", path)
		}
		if entry.IsDir() || !strings.HasSuffix(path, ".templ") {
			return nil
		}
		relative, _ := filepath.Rel(source, path)
		destination := filepath.Join(output, strings.TrimSuffix(strings.ReplaceAll(relative, string(filepath.Separator), "_"), ".templ")+"_templ.go")
		if expected[destination] {
			return fmt.Errorf("generated filename collision: %s", destination)
		}
		expected[destination] = true
		command := exec.Command(strings.TrimSpace(string(tool)), "generate", "-f", path, "-stdout")
		var diagnostics bytes.Buffer
		command.Stderr = &diagnostics
		data, err := command.Output()
		if err != nil {
			return fmt.Errorf("generate %s: %w\n%s", path, err, diagnostics.String())
		}
		data, err = cleanImports(data)
		if err != nil {
			return fmt.Errorf("format %s: %w", path, err)
		}
		// Go's standard marker lets editors and review tools identify generated code.
		banner := fmt.Sprintf("// Code generated from %s by go run ./cmd/generate-templates; DO NOT EDIT.\n// Edit the source template, then regenerate.\n\n", filepath.ToSlash(path))
		data = append([]byte(banner), data...)
		if check {
			existing, err := os.ReadFile(destination)
			if err != nil || !bytes.Equal(existing, data) {
				return fmt.Errorf("%s is stale; run npm run generate:web", destination)
			}
			return nil
		}
		if err := os.MkdirAll(output, 0755); err != nil {
			return err
		}
		return os.WriteFile(destination, data, 0644)
	})
	if err != nil {
		return err
	}
	files, err := os.ReadDir(output)
	if err != nil {
		return err
	}
	for _, file := range files {
		if !strings.HasSuffix(file.Name(), ".go") {
			continue
		}
		path := filepath.Join(output, file.Name())
		if expected[path] {
			continue
		}
		if !strings.HasSuffix(file.Name(), "_templ.go") {
			return fmt.Errorf("%s: handwritten Go does not belong in the generated render package", path)
		}
		if check {
			return fmt.Errorf("orphaned generated file: %s", path)
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !bytes.HasPrefix(data, []byte("// Code generated from ")) {
			return fmt.Errorf("refusing to delete unrecognized file: %s", path)
		}
		if err := os.Remove(path); err != nil {
			return err
		}
	}
	return nil
}

// Template files can share imports across markup and Go helpers. Keep only the
// imports referenced in this output file; source templates remain unchanged.
func cleanImports(data []byte) ([]byte, error) {
	set := token.NewFileSet()
	file, err := parser.ParseFile(set, "generated.go", data, parser.ParseComments)
	if err != nil {
		return nil, err
	}
	used := map[string]bool{}
	ast.Inspect(file, func(node ast.Node) bool {
		if selector, ok := node.(*ast.SelectorExpr); ok {
			if name, ok := selector.X.(*ast.Ident); ok {
				used[name.Name] = true
			}
		}
		return true
	})
	var declarations []ast.Decl
	seenImports := map[string]bool{}
	for _, declaration := range file.Decls {
		group, ok := declaration.(*ast.GenDecl)
		if !ok || group.Tok != token.IMPORT {
			declarations = append(declarations, declaration)
			continue
		}
		var imports []ast.Spec
		for _, spec := range group.Specs {
			value := spec.(*ast.ImportSpec)
			path, _ := strconv.Unquote(value.Path.Value)
			name := filepath.Base(path)
			if value.Name != nil {
				name = value.Name.Name
			}
			if used[name] || name == "_" || name == "." {
				key := name + "\n" + path
				if seenImports[key] {
					continue
				}
				seenImports[key] = true
				imports = append(imports, spec)
			}
		}
		group.Specs = imports
		if len(imports) > 0 {
			declarations = append(declarations, group)
		}
	}
	file.Decls = declarations
	var output bytes.Buffer
	err = format.Node(&output, set, file)
	return output.Bytes(), err
}
