package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/sojupie/tokens/source/dtcg"
	"github.com/sojupie/tokens/source/publish"
)

func runExport(args []string) error {
	flags := flag.NewFlagSet("tokens export", flag.ContinueOnError)
	published, output := "published", "dist/tokens"
	flags.StringVar(&published, "published", published, "publication directory")
	flags.StringVar(&output, "output", output, "output directory for tokens.tokens.json and report.json")
	themes := false
	flags.BoolVar(&themes, "themes", false, "export the site's six appearance contexts and an index")
	context := dtcg.DefaultContext()
	flags.Func("context", "context selection, e.g. 'Theme=Dark'; repeat for each override", func(value string) error {
		key, option, ok := strings.Cut(value, "=")
		if !ok || strings.TrimSpace(key) == "" || strings.TrimSpace(option) == "" {
			return fmt.Errorf("context must be Name=Value")
		}
		context[strings.TrimSpace(key)] = strings.TrimSpace(option)
		return nil
	})
	if err := flags.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return nil
		}
		return err
	}
	if flags.NArg() != 0 {
		return errors.New("export accepts flags only")
	}
	dir, err := publish.ResolveGeneration(published)
	if err != nil {
		return err
	}
	state, err := publish.Load(filepath.Join(dir, "tokens.sqlite"))
	if err != nil {
		return err
	}
	if themes {
		var index []map[string]string
		for _, theme := range []string{"Light", "Dark"} {
			for _, contrast := range []string{"Default", "Medium contrast", "High contrast"} {
				selection := map[string]string{}
				for k, v := range context {
					selection[k] = v
				}
				selection["Theme"], selection["Contrast"] = theme, contrast
				name := strings.ToLower(theme + "-" + strings.TrimSuffix(contrast, " contrast"))
				if err := writeDTCGExport(state, selection, filepath.Join(output, name)); err != nil {
					return err
				}
				index = append(index, map[string]string{"name": name, "theme": strings.ToLower(theme), "contrast": strings.ToLower(strings.TrimSuffix(contrast, " contrast")), "file": name + "/tokens.tokens.json"})
			}
		}
		data, err := json.MarshalIndent(index, "", "  ")
		if err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(output, "index.json"), append(data, '\n'), 0644)
	}
	return writeDTCGExport(state, context, output)
}

func writeDTCGExport(state publish.PublishedState, context map[string]string, output string) error {
	document, report, err := dtcg.Export(state, context)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(output, 0755); err != nil {
		return err
	}
	for name, value := range map[string]any{"tokens.tokens.json": document, "report.json": report} {
		data, err := json.MarshalIndent(value, "", "  ")
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(output, name), append(data, '\n'), 0644); err != nil {
			return err
		}
	}
	fmt.Printf("Exported %d DTCG tokens (%d aliases); %d omissions listed in %s\n", report.Exported, report.Aliases, len(report.Omitted), filepath.Join(output, "report.json"))
	return nil
}
