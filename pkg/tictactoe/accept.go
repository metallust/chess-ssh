package tictactoe

type acceptPage struct {
	question string
	answer   []string
	cursor   int
}

func (a *acceptPage) moveCursor(i int) {
	a.cursor += i
	if a.cursor < 0 {
		a.cursor = 0
	}
	if a.cursor >= len(a.answer) {
		a.cursor = len(a.answer) - 1
	}
}
