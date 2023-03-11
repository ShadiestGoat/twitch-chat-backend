package colors

import (
	"strconv"
)

func HexToRGB(s string) (r, g, b uint8) {
	if len(s) != 6 {
		return
	}

	r = parseHex(s[:2])
	g = parseHex(s[2:4])
	b = parseHex(s[4:])

	return
}

func parseHex(inp string) uint8 {
	p, err := strconv.ParseUint(inp, 16, 8)
	if err != nil {
		panic(err)
	}
	return uint8(p)
}
