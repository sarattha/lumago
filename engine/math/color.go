package math

type Color struct {
	R, G, B, A float32
}

func White() Color {
	return Color{R: 1, G: 1, B: 1, A: 1}
}

func Black() Color {
	return Color{R: 0, G: 0, B: 0, A: 1}
}
