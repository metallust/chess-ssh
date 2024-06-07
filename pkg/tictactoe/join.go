package tictactoe

type joinPage struct {
	Cursor       int
	AvaibleGames []string
}

func (j *joinPage) moveCursor(i int) {
    j.Cursor += i
    if j.Cursor < 0 {
        j.Cursor = 0
    }
    if j.Cursor >= len(j.AvaibleGames) {
        j.Cursor = len(j.AvaibleGames) - 1
    }
}
