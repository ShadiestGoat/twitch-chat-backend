package colors

import (
	"github.com/ShadiestGoat/colorutils"
)

func FixColor(inp string) string {
	r, g, b := HexToRGB(inp)

	h, s, v := colorutils.RGBToHSV(r, g, b)
	h = FixHue(h)

	if s > 0.9 && v > 0.9 {
		return inp
	}

	return colorutils.Hexadecimal(colorutils.HSVToRGB(h, 1, 1))
}
