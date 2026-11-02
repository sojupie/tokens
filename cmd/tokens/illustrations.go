package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	imageformat "image"
	"image/jpeg"
	_ "image/png"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/chromedp"
	"github.com/sojupie/tokens/catalog/web"
	"github.com/sojupie/tokens/source/publish"
)

// runIllustrations renders offline build assets. The serving process never
// launches a browser or generates screenshots on a visitor request.
func runIllustrations(args []string) error {
	flags := flag.NewFlagSet("tokens illustrations", flag.ContinueOnError)
	published := flags.String("published", "published", "publication directory")
	assets := flags.String("assets", "", "generated asset directory; defaults to publication/assets")
	captures := flags.String("from-captures", "", "use existing 720×450 thumbnail captures instead of launching Chromium")
	chrome := flags.String("chrome-path", "", "optional Chromium executable")
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return fmt.Errorf("illustrations does not accept positional arguments")
	}
	if *assets == "" {
		*assets = filepath.Join(*published, "assets")
	}
	state, err := publish.Load(filepath.Join(*published, "tokens.sqlite"))
	if err != nil {
		return err
	}
	server := web.New(state, filepath.Join(*published, "downloads"))
	if err = server.UseGeneratedAssetsDirectory(*published, *assets); err != nil {
		return err
	}
	manifest := server.IllustrationJobs()
	output := filepath.Join(*assets, "illustrations")
	if err = os.MkdirAll(output, 0755); err != nil {
		return err
	}
	// Old content-addressed files remain valid for already open pages. Publish the
	// new manifest only after every subject and appearance has rendered successfully.
	var origin string
	var browser context.Context
	if *captures == "" {
		local := httptest.NewServer(server.Handler())
		defer local.Close()
		origin = local.URL
		options := append([]chromedp.ExecAllocatorOption{}, chromedp.DefaultExecAllocatorOptions[:]...)
		if *chrome != "" {
			options = append(options, chromedp.ExecPath(*chrome))
		}
		allocator, cancel := chromedp.NewExecAllocator(context.Background(), options...)
		defer cancel()
		browser, cancel = chromedp.NewContext(allocator)
		defer cancel()
		// Allocate on the batch context so canceling one capture keeps Chromium alive.
		if err = chromedp.Run(browser); err != nil {
			return fmt.Errorf("start thumbnail renderer: %w", err)
		}
	}
	for i := range manifest.Images {
		image := &manifest.Images[i]
		image.Files = map[string]string{}
		image.Thumbnails = map[string]string{}

		width, height := 720, 450
		for _, theme := range []string{"light", "dark"} {
			for _, contrast := range []string{"default", "medium", "high"} {
				variant := theme + "-" + contrast
				name := strings.Trim(strings.ReplaceAll(image.Path, "/", "-"), "-")
				if name == "" {
					name = "catalog"
				}
				assetName := name + "-thumbnail"
				captureName := assetName + "--" + variant + ".png"
				var capture []byte
				if *captures != "" {
					capture, err = os.ReadFile(filepath.Join(*captures, captureName))
				} else {
					ctx, cancel := context.WithTimeout(browser, 30*time.Second)
					renderURL := origin + "/illustrations/render" + image.Path + "?theme=" + url.QueryEscape(theme) + "&contrast=" + url.QueryEscape(contrast) + "&composition=thumbnail"
					err = chromedp.Run(ctx, chromedp.EmulateViewport(int64(width), int64(height)), chromedp.Navigate(renderURL), chromedp.WaitReady("[data-scene-ready]"), chromedp.CaptureScreenshot(&capture))
					cancel()
				}
				if err != nil {
					return fmt.Errorf("%s / %s: %w", image.Path, variant, err)
				}
				bitmap, _, err := imageformat.Decode(bytes.NewReader(capture))
				if err != nil {
					return err
				}
				if bitmap.Bounds().Dx() != width || bitmap.Bounds().Dy() != height {
					return fmt.Errorf("wrong capture dimensions for %s", captureName)
				}
				var encoded bytes.Buffer
				if err = jpeg.Encode(&encoded, bitmap, &jpeg.Options{Quality: 90}); err != nil {
					return err
				}
				hash := sha256.Sum256(encoded.Bytes())
				file := fmt.Sprintf("%s-%s-%x.jpg", assetName, variant, hash[:8])
				if err = os.WriteFile(filepath.Join(output, file), encoded.Bytes(), 0644); err != nil {
					return err
				}
				image.Thumbnails[variant] = file
			}
		}

		fmt.Printf("%d/%d %s\n", i+1, len(manifest.Images), image.Path)
	}
	data, err := json.MarshalIndent(manifest, "", "  ")
	if err != nil {
		return err
	}
	temporary := filepath.Join(output, "manifest.json.tmp")
	if err = os.WriteFile(temporary, append(data, '\n'), 0644); err != nil {
		return err
	}
	return os.Rename(temporary, filepath.Join(output, "manifest.json"))
}
