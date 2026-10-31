package web

func (c componentCase) ChipLabel() string {
	if c.illustrationLabels[0] != "" {
		return c.illustrationLabels[0]
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
