package presentation

func (c ComponentCase) ChipLabel() string {
	if c.IllustrationLabels[0] != "" {
		return c.IllustrationLabels[0]
	}
	switch c.Variant {
	case "assist":
		return "Assist chip"
	case "suggestion":
		return "Suggestion chip"
	case "input":
		return "Alex"
	default:
		return "Filter chip"
	}
}
