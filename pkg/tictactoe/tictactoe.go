package tictactoe

import (
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/metallust/dosssh/client"
	"github.com/metallust/dosssh/connector"
)

type Model struct {
	Page           int
	User           string
	Opponent       string
	OpponentStatus string
	JoinPage       joinPage
	Game           Tictactoe
	gameClient     *client.GameClient
	ErrorPage      errorPage
	acceptPage     acceptPage
	quit           bool
}

type joinPage struct {
	Cursor       int
	AvaibleGames []string
}

type errorPage struct {
	errorMsg string
}

type acceptPage struct {
	question string
	answer   []string
	cursor   int
}

type Tictactoe struct {
	Board         [3][3]string
	Cursor        [2]int
	Player        string
	CurrentPlayer string
}

const (
	MENUPAGE int = iota
	JOINPAGE
	GAMEPAGE
	LOADINGPAGE
	ERRORPAGE
	ACCEPTPAGE
)

func InitialModel(user string, conn *connector.Connector) tea.Model {
	m := Model{}
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			m.Game.Board[i][j] = " "
		}
	}
	m.gameClient = client.NewGameClient(conn, user)

	m.User = user
	m.Page = MENUPAGE
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
			m.quit = true
			return m, tea.Quit
		}
		if msgString == "m" || msgString == "esc" {
			m.Page = MENUPAGE
			return m, nil
		}

		if m.Page == MENUPAGE {
			switch msgString {
			case "n", "N":
				m.Page = LOADINGPAGE
				return m, m.gameClient.Create("create")
			case "o", "O":
				m.Page = JOINPAGE
				return m, m.gameClient.List("list")
			}
		} else if m.Page == ACCEPTPAGE {
			switch msgString {
			case "up", "k":
				if m.acceptPage.cursor < len(m.acceptPage.answer)-1 {
					m.acceptPage.cursor += 1
				}
			case "down", "j":
				if m.acceptPage.cursor > 0 {
					m.acceptPage.cursor -= 1
				}
			case "enter":
				if m.acceptPage.cursor == 0 {
					return m, m.gameClient.AcceptRequest(true, "joined")
				} else {
					return m, m.gameClient.AcceptRequest(false, "rejected")
				}
			}
		} else if m.Page == JOINPAGE {
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
				m.Page = GAMEPAGE
				m.Opponent = opponent
				m.OpponentStatus = "Requesting ..."
				return m, m.gameClient.Join(opponent, "joined")
			}
		} else if m.Page == GAMEPAGE {
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
					return m, m.gameClient.Move(m.Game.Cursor, "move")
				}
			case "r":
				m.Game.Board = [3][3]string{{" ", " ", " "}, {" ", " ", " "}, {" ", " ", " "}}
				//TODO: add restart feat
			}
		}
	case client.DoneMsg:
		switch msg.Msg {
		case "list":
			m.Page = JOINPAGE
			m.JoinPage.AvaibleGames = msg.Data.([]string)
		case "create":
			m.Page = GAMEPAGE
			m.OpponentStatus = "Waiting ..."
			return m, m.gameClient.ListenServer()
        case "joined":
            m.Page = GAMEPAGE
		    data := msg.Data.([]string)
			m.OpponentStatus = "Connected"
			m.Opponent = data[0]
			if data[1] == "first" {
				m.Game.Player = "X"
			} else {
				m.Game.Player = "O"
				return m, m.gameClient.ListenOpponent()
			}
 		case "rejected":
			m.Page = GAMEPAGE
			return m, m.gameClient.ListenServer()
		case "move":
			m.Game.Board[m.Game.Cursor[0]][m.Game.Cursor[1]] = m.Game.Player
			if m.Game.CurrentPlayer == "X" {
				m.Game.CurrentPlayer = "O"
			} else {
				m.Game.CurrentPlayer = "X"
			}
			return m, m.gameClient.ListenOpponent()
		case "errortimeup":
			m.ErrorPage.errorMsg = ""
			m.Page = MENUPAGE
		}

	case client.GameClientMsg:
		switch msg.Msg {
		case client.JOINREQMSG:
			m.Page = ACCEPTPAGE
			// Opponent := msg.Data.(string)
			Opponent := msg.Data.(string)
			m.acceptPage.question = "Accept request from " + Opponent + " ?"
			m.acceptPage.answer = []string{"Yes", "No"}
			m.acceptPage.cursor = 0
		case client.ERRORMSG:
			m.ErrorPage.errorMsg = msg.Data.(string)
			m.Page = ERRORPAGE
			return m, doTick("errortimeup")
		}
	case client.GameClientOpponentMsg:
		switch msg.Msg {
		case client.DISCONNECTEDMSG:
			m.ErrorPage.errorMsg = "Opponent Disconnected"
			return m, doTick("errortimeup")
		case client.MOVEMSG:
			move := msg.Data.([2]int)
			m.Game.Board[move[0]][move[1]] = m.Game.CurrentPlayer
			m.Game.CurrentPlayer = m.Game.Player
		}
	}
	return m, nil
}

func (m Model) View() string {
	if m.quit {
		return "Chessssh ...bye"
	}

	s := "\n\n\t\tTic Tac Toe"
	switch m.Page {
	// Page 1 will contain Start screen
	case MENUPAGE:
		s += "\n\n"
		s += "\t\tO  Join   Game\n"
		s += "\t\tN  Create Game\n"
		s += "\t\tQ  Quit\n"
		s += "\n\n\t\tQ: quit  r:restart\n"
		return s
		//Page 2 will content the Game
	case GAMEPAGE:
		s += "\n\nMe: " + m.User + "\n"
		s += "Opponent : " + m.Opponent + "    [" + m.OpponentStatus + "]\n"
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
	case JOINPAGE:
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

	case LOADINGPAGE:
		s += "\n\n\n\n"
		s += "\t\tLoading ..."
		s += "\n\n\n\n"
		s += "\t\tQ Quit\n"
		return s

	case ERRORPAGE:
		s += "\n\n\t\t : ERROR : \n\n"
		s += "\t\t" + m.ErrorPage.errorMsg
		s += "\n\n\n\n"
		s += "\t\tQ Quit\n"
		return s
	case ACCEPTPAGE:
		s += "\n\n\t\t : ACCEPTED : \n\n"
		s += "\t\t" + m.acceptPage.question + "\n\n"
		for i, ans := range m.acceptPage.answer {
            if i == m.acceptPage.cursor {
                s += "\t\t>"
            }else {
                s += "\t\t "
            }
            s += ans + "\n"
		}
        return s
	}

	return s
}

func doTick(doneMsg string) tea.Cmd {
	return tea.Tick(time.Second*5, func(t time.Time) tea.Msg {
		return client.DoneMsg{
			Msg:  doneMsg,
			Data: nil,
		}
	})
}
