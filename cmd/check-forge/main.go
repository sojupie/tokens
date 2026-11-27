// check-forge runs explicitly selected browser or captured-publication checks.
package main

import (
	"archive/tar"
	"bytes"
	"compress/gzip"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/chromedp/cdproto/input"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
)

func main() {
	suite := flag.String("suite", "", "react, render or captured")
	artifact := flag.String("package", "", "packed React .tgz (react suite)")
	catalog := flag.String("catalog", "", "running Catalog URL (render suite)")
	checks := flag.String("check", "", "comma-separated conformance page names (render suite)")
	publication := flag.String("publication", "", "publication directory (captured suite)")
	assets := flag.String("assets", "", "built assets directory (captured suite)")
	chrome := flag.String("chrome-path", "", "headless Chromium executable")
	flag.Parse()
	var err error
	switch *suite {
	case "captured":
		err = captured(*publication, *assets)
	case "react", "render":
		err = browser(*suite, *artifact, *catalog, *checks, *chrome)
	default:
		err = errors.New("choose --suite react, render or captured; each suite requires its own inputs")
	}
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func command(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	return cmd.Run()
}

func captured(publication, assets string) error {
	if publication == "" || assets == "" {
		return errors.New("captured checks require --publication and --assets")
	}
	var source, built struct {
		PublicationID string `json:"publicationId"`
	}
	for file, dest := range map[string]any{filepath.Join(publication, "publication.json"): &source, filepath.Join(assets, "manifest.json"): &built} {
		data, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		if err = json.Unmarshal(data, dest); err != nil {
			return err
		}
	}
	if source.PublicationID == "" || source.PublicationID != built.PublicationID {
		return errors.New("publication and built asset identities must match")
	}
	absolute, err := filepath.Abs(assets)
	if err != nil {
		return err
	}
	fmt.Println("Captured publication:", source.PublicationID, "assets:", absolute)
	cmd := exec.Command("node", "--test", "--test-reporter=tap", "--test-name-pattern", "six real themes|captured components|real button sets|disabled text and outlined", "forge/build-tokens.test.mjs", "forge/material/components/generate-components.test.mjs", "forge/material/components/generate-buttons.test.mjs")
	cmd.Env = append(os.Environ(), "TOKEN_ASSETS_DIR="+absolute, "TOKEN_ASSETS_REQUIRED=1")
	var output bytes.Buffer
	cmd.Stdout, cmd.Stderr = io.MultiWriter(os.Stdout, &output), os.Stderr
	if err := cmd.Run(); err != nil {
		return err
	}
	return capturedSummary(output.String())
}

func capturedSummary(output string) error {
	var passed, skipped int
	for _, line := range strings.Split(output, "\n") {
		if strings.HasPrefix(line, "# pass ") {
			_, _ = fmt.Sscanf(line, "# pass %d", &passed)
		}
		if strings.HasPrefix(line, "# skipped ") {
			_, _ = fmt.Sscanf(line, "# skipped %d", &skipped)
		}
	}
	if passed != 4 || skipped != 0 {
		return fmt.Errorf("required captured checks: expected four passes and no skips, got %d passes and %d skips", passed, skipped)
	}
	return nil
}

func unpack(artifact, destination string) error {
	file, err := os.Open(artifact)
	if err != nil {
		return err
	}
	defer file.Close()
	zip, err := gzip.NewReader(file)
	if err != nil {
		return err
	}
	defer zip.Close()
	reader := tar.NewReader(zip)
	for {
		header, err := reader.Next()
		if err == io.EOF {
			return nil
		}
		if err != nil {
			return err
		}
		name := strings.ReplaceAll(header.Name, "\\", "/")
		if !strings.HasPrefix(name, "package/") || !filepath.IsLocal(filepath.FromSlash(name)) {
			return fmt.Errorf("invalid package path: %q", name)
		}
		target := filepath.Join(destination, filepath.FromSlash(name))
		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return err
			}
		case tar.TypeReg:
			if err := os.MkdirAll(filepath.Dir(target), 0755); err != nil {
				return err
			}
			out, err := os.Create(target)
			if err != nil {
				return err
			}
			_, err = io.Copy(out, reader)
			closeErr := out.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
		default:
			return fmt.Errorf("unsupported package entry: %q", name)
		}
	}
}

func browser(suite, artifact, catalog, checks, chrome string) error {
	var handler http.Handler
	if suite == "react" {
		if artifact == "" {
			return errors.New("react checks require --package pointing to the packed .tgz")
		}
		if err := os.MkdirAll("tmp", 0755); err != nil {
			return err
		}
		dir, err := os.MkdirTemp("tmp", "forge-check-")
		if err != nil {
			return err
		}
		// Retain this repo-scoped fixture for inspection; never remove arbitrary package paths.
		absolute, err := filepath.Abs(dir)
		if err != nil {
			return err
		}
		if err = unpack(artifact, absolute); err != nil {
			return err
		}
		manifest, err := os.ReadFile(filepath.Join(absolute, "package", "package.json"))
		if err != nil {
			return err
		}
		var pkg struct{ Name, Version string }
		if err = json.Unmarshal(manifest, &pkg); err != nil {
			return err
		}
		if pkg.Name != "@token-catalog/forge-react" {
			return errors.New("expected the Forge React package")
		}
		fmt.Printf("Packed consumer: %s %s (%s)\n", pkg.Name, pkg.Version, artifact)
		archive, err := os.ReadFile(artifact)
		if err != nil {
			return err
		}
		fmt.Printf("Package SHA-256: %x\n", sha256.Sum256(archive))
		out := filepath.Join(absolute, "site")
		if err = command("node", "forge/react/tests/build.ts", filepath.Join(absolute, "package"), out); err != nil {
			return err
		}
		handler = http.FileServer(http.Dir(out))
	} else {
		target, err := url.Parse(catalog)
		if err != nil || target.Host == "" || (target.Scheme != "http" && target.Scheme != "https") || checks == "" {
			return errors.New("render checks require --catalog URL and --check page-name[,page-name]")
		}
		for _, name := range strings.Split(checks, ",") {
			if !filepath.IsLocal(name) || strings.ContainsAny(name, "/\\.") {
				return fmt.Errorf("invalid check name %q", name)
			}
		}
		proxy := httputil.NewSingleHostReverseProxy(target)
		handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			switch {
			case strings.HasSuffix(r.URL.Path, ".html"):
				http.FileServer(http.Dir("forge/conformance/pages")).ServeHTTP(w, r)
			case strings.HasSuffix(r.URL.Path, ".mjs"), strings.HasSuffix(r.URL.Path, "-check.js"):
				http.FileServer(http.Dir("forge/conformance/browser")).ServeHTTP(w, r)
			default:
				r.URL.Path = strings.TrimPrefix(r.URL.Path, "/md3")
				proxy.ServeHTTP(w, r)
			}
		})
	}
	server := httptest.NewServer(handler)
	defer server.Close()
	opts := append(chromedp.DefaultExecAllocatorOptions[:], chromedp.WindowSize(1280, 1000))
	if chrome != "" {
		opts = append(opts, chromedp.ExecPath(chrome))
	}
	allocator, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()
	ctx, cancel := chromedp.NewContext(allocator)
	defer cancel()
	ctx, cancel = context.WithTimeout(ctx, 3*time.Minute)
	defer cancel()
	var exceptions []string
	var exceptionLock sync.Mutex
	chromedp.ListenTarget(ctx, func(event any) {
		if exception, ok := event.(*runtime.EventExceptionThrown); ok {
			exceptionLock.Lock()
			defer exceptionLock.Unlock()
			exceptions = append(exceptions, exception.ExceptionDetails.Text)
		}
	})
	var err error
	if suite == "react" {
		err = reactChecks(ctx, server.URL)
	} else {
		for _, name := range strings.Split(checks, ",") {
			if err = chromedp.Run(ctx, chromedp.Navigate(server.URL+"/"+name+".html"), chromedp.Click("#run", chromedp.ByQuery), chromedp.Poll(`(()=>{try{JSON.parse(document.querySelector('#result').textContent);return true}catch{return false}})()`, nil, chromedp.WithPollingTimeout(90*time.Second))); err != nil {
				break
			}
			var result json.RawMessage
			if err = chromedp.Run(ctx, chromedp.Evaluate(`JSON.parse(document.querySelector('#result').textContent)`, &result)); err != nil {
				break
			}
			if err = accept(result); err != nil {
				err = fmt.Errorf("%s: %w", name, err)
				break
			}
			fmt.Printf("PASS %s %s\n", name, result)
		}
	}
	if err != nil {
		return err
	}
	exceptionLock.Lock()
	defer exceptionLock.Unlock()
	if len(exceptions) > 0 {
		return fmt.Errorf("browser exceptions: %v", exceptions)
	}
	return nil
}

// Missing/empty results are never evidence of conformance.
func accept(data []byte) error {
	var result struct {
		Passed, Checked, PropertiesChecked, FailureCount int
		Error                                            string
		Failures                                         []json.RawMessage
	}
	if err := json.Unmarshal(data, &result); err != nil {
		return err
	}
	if result.Error != "" || result.FailureCount > 0 || len(result.Failures) > 0 || result.Passed+result.Checked+result.PropertiesChecked <= 0 {
		return fmt.Errorf("check failed: %s", data)
	}
	return nil
}

func evaluate(ctx context.Context, expression string, result any) error {
	return chromedp.Run(ctx, chromedp.Evaluate(expression, result, func(p *runtime.EvaluateParams) *runtime.EvaluateParams { return p.WithAwaitPromise(true) }))
}

func assertJS(ctx context.Context, expression, message string) error {
	var ok bool
	if err := evaluate(ctx, expression, &ok); err != nil {
		return err
	}
	if !ok {
		return errors.New(message)
	}
	return nil
}

func poll(ctx context.Context, expression string) error {
	return chromedp.Run(ctx, chromedp.Poll(expression, nil, chromedp.WithPollingTimeout(5*time.Second)))
}

func reactChecks(ctx context.Context, base string) error {
	if err := chromedp.Run(ctx, chromedp.Navigate(base), chromedp.WaitVisible("#open-outer", chromedp.ByQuery)); err != nil {
		return err
	}
	if err := evaluate(ctx, `document.fonts.ready.then(()=>true)`, nil); err != nil {
		return err
	}
	if err := chromedp.Run(ctx, chromedp.Click("#open-outer", chromedp.ByQuery)); err != nil {
		return err
	}
	if err := poll(ctx, `document.querySelector('#outer-dialog').open`); err != nil {
		return err
	}
	if err := evaluate(ctx, `document.querySelector('#outer-draft').value='Edited draft'`, nil); err != nil {
		return err
	}
	for index, mode := range []string{"action", "escape", "controlled"} {
		if err := chromedp.Run(ctx, chromedp.Click("#open-inner", chromedp.ByQuery)); err != nil {
			return err
		}
		if err := poll(ctx, `document.querySelector('#inner-dialog').open`); err != nil {
			return err
		}
		var err error
		switch mode {
		case "escape":
			err = chromedp.Run(ctx, chromedp.KeyEvent("\u001b"))
		case "controlled":
			err = chromedp.Run(ctx, chromedp.Click("#controlled-inner", chromedp.ByQuery))
		default:
			err = chromedp.Run(ctx, chromedp.Click("#close-inner", chromedp.ByQuery))
		}
		if err != nil {
			return err
		}
		if err = poll(ctx, fmt.Sprintf(`JSON.parse(document.querySelector('#dialog-state').textContent).innerClosed===%d`, index+1)); err != nil {
			return fmt.Errorf("%s child close: %w", mode, err)
		}
		if err = assertJS(ctx, `(()=>{const s=JSON.parse(document.querySelector('#dialog-state').textContent);return s.outer&&!s.inner&&s.closed===0&&document.querySelector('#outer-dialog').open&&document.querySelector('#outer-draft').value==='Edited draft'&&document.activeElement.id==='open-inner'})()`, "child dismissal closed its parent, lost its draft or returned focus incorrectly"); err != nil {
			return err
		}
	}
	if err := chromedp.Run(ctx, chromedp.Click("#close-outer", chromedp.ByQuery)); err != nil {
		return err
	}
	if err := poll(ctx, `JSON.parse(document.querySelector('#dialog-state').textContent).closed===1&&document.activeElement.id==='open-outer'`); err != nil {
		return err
	}
	fmt.Println("PASS packed React nested dialogs: action, Escape, controlled close, draft retention, callbacks and focus return")
	if err := chromedp.Run(ctx, chromedp.Focus("#field", chromedp.ByQuery)); err != nil {
		return err
	}
	if err := evaluate(ctx, `new Promise(requestAnimationFrame).then(()=>Promise.all(document.querySelector('#field-fixture').getAnimations({subtree:true}).map(a=>a.finished.catch(()=>{}))))`, nil); err != nil {
		return err
	}
	var layout json.RawMessage
	if err := evaluate(ctx, `import('./rendered-checks.mjs').then(m=>m.checkLayout())`, &layout); err != nil {
		return err
	}
	if err := accept(layout); err != nil {
		return err
	}
	fmt.Printf("PASS packed layout %s\n", layout)
	if err := assertJS(ctx, `import('./rendered-checks.mjs').then(m=>{const action=document.querySelector('#row-action');action.style.translate='-200px';const failed=m.checkLayout().failureCount>0;action.style.translate='';return failed})`, "layout negative control did not detect overlapping action"); err != nil {
		return err
	}
	if err := evaluate(ctx, `import('./rendered-checks.mjs').then(m=>m.startGlyph())`, nil); err != nil {
		return err
	}
	var point struct{ X, Y float64 }
	if err := evaluate(ctx, `(()=>{const r=document.querySelector('#glyph-item').getBoundingClientRect();return {x:r.x+r.width/2,y:r.y+r.height/2}})()`, &point); err != nil {
		return err
	}
	for _, phase := range []string{"hover", "press", "release", "leave"} {
		var action chromedp.Action
		switch phase {
		case "hover":
			action = chromedp.MouseEvent(input.MouseMoved, point.X, point.Y)
		case "press":
			action = chromedp.MouseEvent(input.MousePressed, point.X, point.Y, chromedp.ButtonLeft, chromedp.ClickCount(1))
		case "release":
			action = chromedp.MouseEvent(input.MouseReleased, point.X, point.Y, chromedp.ButtonLeft, chromedp.ClickCount(1))
		default:
			action = chromedp.MouseEvent(input.MouseMoved, 1200, 900)
		}
		if err := chromedp.Run(ctx, action); err != nil {
			return err
		}
		var result json.RawMessage
		if err := evaluate(ctx, `import('./rendered-checks.mjs').then(m=>m.checkGlyph('`+phase+`'))`, &result); err != nil {
			return err
		}
		if err := accept(result); err != nil {
			return err
		}
		fmt.Printf("PASS glyph %s %s\n", phase, result)
	}
	for _, selector := range []string{"#split-primary", "#split-menu"} {
		if err := evaluate(ctx, `(()=>{const r=document.querySelector('`+selector+`').getBoundingClientRect();return {x:r.x+r.width/2,y:r.y+r.height/2}})()`, &point); err != nil {
			return err
		}
		for _, phase := range []string{"hover", "press", "release", "leave"} {
			var action chromedp.Action
			switch phase {
			case "hover":
				action = chromedp.MouseEvent(input.MouseMoved, point.X, point.Y)
			case "press":
				action = chromedp.MouseEvent(input.MousePressed, point.X, point.Y, chromedp.ButtonLeft, chromedp.ClickCount(1))
			case "release":
				action = chromedp.MouseEvent(input.MouseReleased, point.X, point.Y, chromedp.ButtonLeft, chromedp.ClickCount(1))
			default:
				action = chromedp.MouseEvent(input.MouseMoved, 1200, 900)
			}
			if err := chromedp.Run(ctx, action); err != nil {
				return err
			}
			var result json.RawMessage
			if err := evaluate(ctx, `import('./rendered-checks.mjs').then(m=>m.checkSplit())`, &result); err != nil {
				return err
			}
			if err := accept(result); err != nil {
				return fmt.Errorf("%s %s: %w", selector, phase, err)
			}
		}
	}
	fmt.Println("PASS split buttons: source corners through hover, press, release, selected and pointer leave")
	return nil
}
