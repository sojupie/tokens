package web

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
)

// UseGeneratedAssets loads Style Dictionary output belonging to this exact
// publication. Publications created before the exporter keep their Go theme.
func (server *Server) UseGeneratedAssets(publicationDir string) error {
	directory := filepath.Join(publicationDir, "assets")
	data, err := os.ReadFile(filepath.Join(directory, "manifest.json"))
	if os.IsNotExist(err) {
		return nil
	}
	if err != nil {
		return err
	}
	var manifest struct{ PublicationID, SnapshotID, DatabaseSha256, CssSha256, ButtonsSha256, ComponentsSha256 string }
	if err := json.Unmarshal(data, &manifest); err != nil {
		return err
	}
	if manifest.PublicationID != server.state.PublicationID || manifest.SnapshotID != server.state.CurrentSnapshot.ID {
		return fmt.Errorf("generated token assets belong to another publication")
	}
	file, err := os.Open(filepath.Join(publicationDir, "tokens.sqlite"))
	if err != nil {
		return err
	}
	hash := sha256.New()
	_, copyErr := io.Copy(hash, file)
	closeErr := file.Close()
	if copyErr != nil {
		return copyErr
	}
	if closeErr != nil {
		return closeErr
	}
	if hex.EncodeToString(hash.Sum(nil)) != manifest.DatabaseSha256 {
		return fmt.Errorf("generated token assets are stale; run npm run build:tokens -- --published %s", publicationDir)
	}
	css, err := os.ReadFile(filepath.Join(directory, "theme.css"))
	if err != nil {
		return err
	}
	sum := sha256.Sum256(css)
	if hex.EncodeToString(sum[:]) != manifest.CssSha256 {
		return fmt.Errorf("generated theme checksum mismatch")
	}
	buttons, err := readStaticFile("static/buttons.css")
	if err != nil {
		return err
	}
	buttonSum := sha256.Sum256(buttons)
	if hex.EncodeToString(buttonSum[:]) != manifest.ButtonsSha256 {
		return fmt.Errorf("generated button bindings differ from the server binary; rebuild token assets and the server")
	}
	components, err := readStaticFile("static/components.css")
	if err != nil {
		return err
	}
	componentSum := sha256.Sum256(components)
	if hex.EncodeToString(componentSum[:]) != manifest.ComponentsSha256 {
		return fmt.Errorf("generated component bindings differ from the server binary; rebuild token assets and the server")
	}

	server.themeCSS = string(css)
	server.themeError = nil
	server.tokenAssetsDir = directory
	return nil
}

func (server *Server) handleDTCGDownload(response http.ResponseWriter, request *http.Request) {
	response.Header().Set("Cache-Control", "public, no-cache")
	variant, file := request.PathValue("variant"), request.PathValue("file")
	valid := false
	for _, allowed := range []string{"light-default", "light-medium", "light-high", "dark-default", "dark-medium", "dark-high"} {
		if variant == allowed {
			valid = true
		}
	}
	if !valid || server.tokenAssetsDir == "" || (file != "tokens.tokens.json" && file != "tokens.css" && file != "report.json") {
		http.NotFound(response, request)
		return
	}
	if file == "tokens.tokens.json" {
		response.Header().Set("Content-Type", "application/design-tokens+json")
	} else if file == "tokens.css" {
		response.Header().Set("Content-Type", "text/css; charset=utf-8")
	} else {
		response.Header().Set("Content-Type", "application/json")
	}
	http.ServeFile(response, request, filepath.Join(server.tokenAssetsDir, "dtcg", variant, file))
}
