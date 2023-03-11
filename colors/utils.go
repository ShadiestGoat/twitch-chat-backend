package colors

func FixHue(inp int) int {
	if inp > 190 && inp < 300 {
		if inp > 245 {
			inp = 300
		} else {
			inp = 190
		}
	}

	return inp
}
