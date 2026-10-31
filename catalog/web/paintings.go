package web

// Painting choices are explicit to keep each editorial scene stable.
var scenePaintings = map[string][2]string{
	"card":                {"italian-coast-scene-with-ruined-tower-1993.55.1.webp", "swiss-landscape-2003.41.1.webp"},
	"button":              {"cliffs-at-pourville-1985.64.27.webp", "cliffs-at-pourville-1985.64.27.webp"},
	"extended-fab":        {"palmer-river-2021.11.1.webp", "palmer-river-2021.11.1.webp"},
	"fab":                 {"peaceful-valley-1965.10.1.webp", "peaceful-valley-1965.10.1.webp"},
	"split":               {"the-juniata-evening-2010.107.1.webp", "the-juniata-evening-2010.107.1.webp"},
	"chip":                {"autumn---on-the-hudson-river-1963.9.1.webp", "argenteuil-1970.17.42.webp"},
	"segmented":           {"pastoral-landscape-1991.96.1.webp", "beach-at-trouville-1970.17.12.webp"},
	"button-group":        {"green-river-cliffs-wyoming-2011.2.1.webp", "sunset-at-scheveningen-a-fleet-of-fishing-vessels-at-anchor-2012.48.1.webp"},
	"tabs":                {"landscape-distant-village-2015.19.10.webp", "le-havre-2015.19.58.webp"},
	"slider":              {"giant-magnolias-on-a-blue-velvet-cloth-1996.14.1.webp", "giant-magnolias-on-a-blue-velvet-cloth-1996.14.1.webp"},
	"nav-drawer":          {"the-island-and-bridge-of-san-bartolomeo-rome-2001.23.1.webp", "the-fortress-of-konigstein-1993.8.1.webp"},
	"nav-rail":            {"nymphenburg-palace-munich-1961.9.63.webp", "place-du-carrousel-paris-1970.17.55.webp"},
	"nav-bar":             {"yarmouth-jetty-2016.115.1.webp", "low-tide-at-scheveningen-2025.11.7.webp"},
	"appbar":              {"washerwomen-on-the-beach-of-etretat-1970.17.17.webp", "lake-lucerne-1990.50.1.webp"},
	"dialog":              {"fanciful-landscape-1963.9.2.webp", "landscape-2014.136.109.webp"},
	"bottom-sheet":        {"the-voyage-of-life-youth-1971.16.2.webp", "blue-morning-1963.10.82.webp"},
	"side-sheet":          {"the-roman-theater-at-taormina-2004.166.33.webp", "the-louvre-afternoon-rainy-weather-2014.136.59.webp"},
	"snackbar":            {"the-grand-canal-2001.87.1.webp", "the-grand-canal-2001.87.1.webp"},
	"toolbar":             {"the-maas-at-dordrecht-1940.2.1.webp", "the-maas-at-dordrecht-1940.2.1.webp"},
	"list":                {"entrance-to-the-grand-canal-from-the-molo-venice-1945.15.4.webp", "entrance-to-the-grand-canal-from-the-molo-venice-1945.15.4.webp"},
	"divider":             {"lake-lucerne-1990.50.1.webp", "lake-lucerne-1990.50.1.webp"},
	"icon":                {"peaceful-valley-1965.10.1.webp", "peaceful-valley-1965.10.1.webp"},
	"shape/corner-radius": {"green-river-cliffs-wyoming-2011.2.1.webp", "giant-magnolias-on-a-blue-velvet-cloth-1996.14.1.webp"},
	"elevation":           {"palmer-river-2021.11.1.webp", "palmer-river-2021.11.1.webp"},
}

var carouselPaintings = [...]string{
	"argenteuil-1970.17.42.webp",
	"autumn---on-the-hudson-river-1963.9.1.webp",
	"cliffs-at-pourville-1985.64.27.webp",
	"green-river-cliffs-wyoming-2011.2.1.webp",
	"lake-lucerne-1990.50.1.webp",
	"sunset-at-scheveningen-a-fleet-of-fishing-vessels-at-anchor-2012.48.1.webp",
}

func scenePainting(scene, slot string) string {
	images, ok := scenePaintings[scene]
	if !ok {
		images = scenePaintings["card"]
	}
	if slot == "hills" {
		return images[1]
	}
	return images[0]
}
