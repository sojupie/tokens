package main

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	imageformat "image"
	"image/color"
	"image/jpeg"
	_ "image/png"
	"net/http/httptest"
	"net/url"
	"os"
	"os/exec"
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
	encoder, err := exec.LookPath("cwebp")
	if err != nil {
		return fmt.Errorf("thumbnail generation requires cwebp: %w", err)
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
		image.SmallThumbnails = map[string]string{}
		image.Placeholders = map[string]string{}
		image.PlaceholderFormat = "webp"

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
				placeholder, err := thumbnailWebP(encoder, capture)
				if err != nil {
					return fmt.Errorf("%s / %s: %w", image.Path, variant, err)
				}
				image.Placeholders[variant] = base64.StdEncoding.EncodeToString(placeholder)
				for _, size := range []struct{ width, height, quality int }{{720, 450, 80}, {360, 225, 75}} {
					encoded, err := thumbnailJPEG(bitmap, size.width, size.height, size.quality)
					if err != nil {
						return err
					}
					hash := sha256.Sum256(encoded)
					file := fmt.Sprintf("%s-%d-%s-%x.jpg", assetName, size.width, variant, hash[:8])
					if err = os.WriteFile(filepath.Join(output, file), encoded, 0644); err != nil {
						return err
					}
					if size.width == 720 {
						image.Thumbnails[variant] = file
					} else {
						image.SmallThumbnails[variant] = file
					}
				}
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

func thumbnailWebP(encoder string, capture []byte) ([]byte, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	cmd := exec.CommandContext(ctx, encoder, "-quiet", "-q", "15", "-m", "6", "-resize", "40", "25", "-o", "-", "--", "-")
	cmd.Stdin = bytes.NewReader(capture)
	output, err := cmd.Output()
	if err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return nil, fmt.Errorf("encode WebP placeholder: %w: %s", err, strings.TrimSpace(string(exit.Stderr)))
		}
		return nil, fmt.Errorf("encode WebP placeholder: %w", err)
	}
	return output, nil
}

func thumbnailJPEG(source imageformat.Image, width, height, quality int) ([]byte, error) {
	bitmap := source
	if source.Bounds().Dx() != width || source.Bounds().Dy() != height {
		resized := imageformat.NewRGBA(imageformat.Rect(0, 0, width, height))
		bounds := source.Bounds()
		for y := 0; y < height; y++ {
			for x := 0; x < width; x++ {
				var red, green, blue, count uint64
				for sy := y * bounds.Dy() / height; sy < (y+1)*bounds.Dy()/height; sy++ {
					for sx := x * bounds.Dx() / width; sx < (x+1)*bounds.Dx()/width; sx++ {
						r, g, b, _ := source.At(bounds.Min.X+sx, bounds.Min.Y+sy).RGBA()
						red += uint64(r)
						green += uint64(g)
						blue += uint64(b)
						count++
					}
				}
				resized.SetRGBA(x, y, color.RGBA{uint8(red / count >> 8), uint8(green / count >> 8), uint8(blue / count >> 8), 255})
			}
		}
		bitmap = resized
	}
	var output bytes.Buffer
	if err := jpeg.Encode(&output, bitmap, &jpeg.Options{Quality: quality}); err != nil {
		return nil, err
	}
	return output.Bytes(), nil
}
