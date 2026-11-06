package presentation

import (
	webassets "github.com/sojupie/tokens/catalog/web/assets"
	"github.com/sojupie/tokens/forge/components"
)

// Catalog versions the geometry asset; Forge owns its interpretation and first frame.
var LoadingGeometryVersion = func() string {
	data, err := components.Files.ReadFile("static/loading-shapes.json")
	if err != nil {
		panic(err)
	}
	return webassets.AssetVersion(data)
}()
