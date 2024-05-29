package tictactoe

import (
	tea "github.com/charmbracelet/bubbletea"
	"github.com/metallust/chessh/internal/connector"
	"github.com/metallust/chessh/internal/game"
)

type Model struct {
	Page           string
	User           string
    Opponent       string
    OpponentStatus string
    JoinPage       joinPage
	Game           Tictactoe
	gameClient     *game.GameClient
}

type joinPage struct {
	Cursor       int
	AvaibleGames []string
}

type Tictactoe struct {
	Board         [3][3]string
	Cursor        [2]int
	Player        string
	CurrentPlayer string
}

func InitialModel(User string, c *connector.Connector) tea.Model {
	m := Model{}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			m.Game.Board[i][j] = " "
		}
	}
    m.gameClient = game.NewGameClient(c)

    m.User = User
    m.Page = "menu"
    m.Game.CurrentPlayer = "X"
	m.Opponent = "Waiting ... "
	return m
}

func (m Model) Init() tea.Cmd {
	return nil
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		msgString := msg.String()
		if msgString == "q" || msgString == "ctrl+c" {
			return m, tea.Quit
		}
		if msgString == "m" || msgString == "esc" {
			m.Page = "menu"
            return m, nil
		}

		if m.Page == "menu" {
			switch msgString {
			case "n", "N":
                //TODO: Loading screen
                m.gameClient.Create()
                m.Page = "game"
                return m, m.gameClient.ListenServer()
			case "o", "O":
                //TODO: Loading screen
                m.JoinPage.AvaibleGames, _ = m.gameClient.List()
    			m.Page = "joinpage"
                return m, nil
			}
		} else if m.Page == "game" {
			switch msgString {
			case "up", "k":
				m.Game.Cursor[0] = (m.Game.Cursor[0] + 2) % 3
				//maybe return a function which emits a custom msg indicating turn
			case "down", "j":
				m.Game.Cursor[0] = (m.Game.Cursor[0] + 1) % 3
			case "left", "h":
				m.Game.Cursor[1] = (m.Game.Cursor[1] + 2) % 3
			case "right", "l":
				m.Game.Cursor[1] = (m.Game.Cursor[1] + 1) % 3
			case " ", "enter":
				if m.Game.Board[m.Game.Cursor[0]][m.Game.Cursor[1]] == " " && m.Game.CurrentPlayer == m.Game.Player {
                    m.gameClient.Move(m.Game.Cursor)
					m.Game.Board[m.Game.Cursor[0]][m.Game.Cursor[1]] = m.Game.Player
					if m.Game.Player == "X" {
						m.Game.CurrentPlayer = "O"
					} else {
						m.Game.CurrentPlayer = "X"
					}
				}
                return m, m.gameClient.ListenServer()
			case "r":
				m.Game.Board = [3][3]string{{" ", " ", " "}, {" ", " ", " "}, {" ", " ", " "}}
                //TODO: add restart feat
			}
		} else if m.Page == "joinpage" {
			switch msgString {
			case "up", "k":
				if m.JoinPage.Cursor < len(m.JoinPage.AvaibleGames)-1 {
					m.JoinPage.Cursor += 1
				}
			case "down", "j":
				if m.JoinPage.Cursor > 0 {
					m.JoinPage.Cursor -= 1
				}
			case "enter":
                if len(m.JoinPage.AvaibleGames) == 0 {
                    return m, nil
                }
				opponent := m.JoinPage.AvaibleGames[m.JoinPage.Cursor]
                //TODO: Loading screen
                first, err :=  m.gameClient.Join(opponent)
                if err != nil {
                    //handle error show error page maybe
                }
                m.Page = "game"
                m.Opponent = opponent
                m.OpponentStatus = "Connected"
                if first {
                    m.Game.Player = "X"
                    return m, nil
                } else {
                    m.Game.Player = "O"
                    return m, m.gameClient.ListenServer()
                }
			}

		}
    case game.GameClientMsg:
        switch msg.Msg {
            case "joined":
                first := msg.Data.(map[string]interface{})["first"].(bool)
                opponentname := msg.Data.(map[string]interface{})["opponent"].(string)
                m.OpponentStatus = "Connected"
                if first {
                    m.Game.Player = "X"
                } else {
                    m.Game.Player = "O"
                }
                m.Opponent = opponentname
            case "move":
                //move
            case "exit":
                //exit
                break

        }
	}
	return m, nil
}

func (m Model) View() string {
	s := "\n\n\t\tTick Tac Toe"
	switch m.Page {
	// Page 1 will contain Start screen
	case "menu":
		s += "\n\n"
		s += "\t\tO    Join   Game\n"
		s += "\t\tN    Create Game\n"
		s += "\t\tQ    Quit\n"
		s += "\n\n\t\tQ: quit  r:restart\n"
		return s
		//Page 2 will content the Game
	case "game":
		s += "\n\nMe: " + m.User + "\n"
		s += "Opponent : " + m.Opponent + "    " + m.OpponentStatus + "\n"
		for y, row := range m.Game.Board {
			if m.Game.Cursor[0] == y {
				s += ">"
			} else {
				s += " "
			}
			for x, cell := range row {
				if m.Game.Cursor[0] == y && m.Game.Cursor[1] == x && cell == " " {
					cell = "."
				}
				s += "|" + cell
			}
			s += "|\n"
		}

		s += " "
		for i := 0; i < 3; i++ {
			if m.Game.Cursor[1] == i {
				s += " ^"
			} else {
				s += "  "
			}
		}

		s += "\nPlayer : " + m.Game.Player
		s += "\nCurent player : " + m.Game.CurrentPlayer
		s += "\nQ: quit  r:restart\n"
		return s
		//List of page that can be Joined
	case "joinpage":
		// s += " - Join Game \n Input Field : " + m.JoinPage.Input + "\n"
		s += "\n"
		for i := range m.JoinPage.AvaibleGames {
			if i == m.JoinPage.Cursor {
				s += "> "
			} else {
				s += "  "
			}
			s += "\t\t" + m.JoinPage.AvaibleGames[i] + "\n"
		}
		s += "\n\n\n\nQ Quit\n"
		return s
	case "loading":
		s += "\n\n\n\n"
		s += "\t\tLoading ..."
		s += "\n\n\n\n"
		s += "\t\tQ Quit\n"
	}
	return s
}
