package presentation

import "fmt"

// Public-domain paintings used by component previews.
const CardPainting = "blue-morning-1963.10.82.webp"

// Impressionist paintings keep the carousel visually coherent.
var CarouselPaintings = [...]string{
	"beach-at-trouville-1970.17.12.webp",
	"woman-with-a-parasol-1983.1.29.webp",
	"argenteuil-1970.17.42.webp",
	"cliffs-at-pourville-1985.64.27.webp",
	"washerwomen-on-the-beach-of-etretat-1970.17.17.webp",
	"the-louvre-afternoon-rainy-weather-2014.136.59.webp",
}

var CarouselPaintingTitles = [...]string{
	"Beach at Trouville", "Woman with a Parasol — Madame Monet and Her Son", "Argenteuil", "Cliffs at Pourville", "Washerwomen on the Beach of Étretat", "The Louvre, Afternoon, Rainy Weather",
}

// Keep people, shorelines and buildings in frame when artwork is cover-cropped.
var CarouselPaintingPositions = [...]string{
	"50% 100%", "52% 10%", "52% 100%", "48% 100%", "50% 100%", "50% 100%",
}

// Reduce empty sky while retaining enough context around each subject.
var CarouselPaintingScales = [...]float64{1.35, 1, 1.15, 1.1, 1.1, 1.05}

func CarouselArtworkStyle(index int) string {
	position := CarouselPaintingPositions[index%len(CarouselPaintingPositions)]
	return fmt.Sprintf("object-position:%s;transform:scale(%g);transform-origin:%s", position, CarouselPaintingScales[index%len(CarouselPaintingScales)], position)
}
