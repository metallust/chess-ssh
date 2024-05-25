package tictactoe

import (
	"log"
	"math/rand"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/metallust/chessh/internal/connector"
)

type Model struct {
	Page           int
	User           string
	Game           Tictactoe
	Opponent       string
	OpponentConn   *connector.Connector
	OpponentStatus string
	JoinPage       Page3
	Connector      *connector.Connector
}

type Page3 struct {
	Input        string
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
	m.Connector = c
	m.Game.CurrentPlayer = "X"
	m.Page = 1
	m.User = User
	m.Opponent = "Waiting ... "
	return m
}

func ListentoServer(m Model) tea.Cmd {
	return func() tea.Msg {
		msg, more := m.Connector.GetMsg()
		if more {
			return msg
		}
		return tea.Quit
	}
}

func ListentoOpponent(m Model) tea.Cmd {
	return func() tea.Msg {
		msg, more := m.OpponentConn.GetMsg()
		if more {
			return msg
		}
		return tea.Quit
	}
}

func (m Model) Init() tea.Cmd {
	return ListentoServer(m)
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		msgString := msg.String()
		if msgString == "Q" || msgString == "ctrl+c" {
			// m.Channel <- ServerMsg{Msg: "exit", Data: nil}
			return m, tea.Quit
		}
		if msgString == "B" {
			m.Page = 1
		}
		// commands for start menu
		if m.Page == 1 {
			switch msgString {
			case "n", "N":
				m.Page = 4
				m.Connector.SendMsg(
					connector.Msg{
						Name: "create",
						Data: nil,
					},
				)
			case "o", "O":
				m.Page = 4
				m.Connector.SendMsg(connector.Msg{
					Name: "list",
					Data: nil,
				})
			}
			// commands for the game
		} else if m.Page == 2 {
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
					m.OpponentConn.SendMsg(connector.Msg{Name: "move", Data: m.Game.Cursor})
					m.Game.Board[m.Game.Cursor[0]][m.Game.Cursor[1]] = m.Game.Player

					if m.Game.Player == "X" {
						m.Game.CurrentPlayer = "O"
					} else {
						m.Game.CurrentPlayer = "X"
					}
				}
			case "r":
				m.Game.Board = [3][3]string{{" ", " ", " "}, {" ", " ", " "}, {" ", " ", " "}}
			}
		} else if m.Page == 3 {
			switch msgString {
			case "up", "k":
				if m.JoinPage.Cursor < len(m.JoinPage.AvaibleGames)-1 {
					m.JoinPage.Cursor += 1
				}
			case "down", "j":
				if m.JoinPage.Cursor > 0 {
					m.JoinPage.Cursor -= 1
				}
			case "backspace":
				if len(m.JoinPage.Input) > 0 {
					// delete 1 item
					m.JoinPage.Input = m.JoinPage.Input[:len(m.JoinPage.Input)-1]
				}
			case "enter":
				// oppChannel := make(chan OpponentMsg)
				// m.Channel <- ServerMsg{
				// Msg: "connect",
				// Data: map[string]interface{}{
				// "channel": oppChannel,
				// },
				// }

				opponent := m.JoinPage.AvaibleGames[m.JoinPage.Cursor]
				m.Connector.SendMsg(connector.Msg{
					Name: "join",
					Data: opponent,
				})
				m.Opponent = opponent
				m.OpponentStatus = "Requested to join ...."
				m.Page = 2
			default:
				m.JoinPage.Input += msgString
			}

		}
	case connector.Msg:
		switch msg.Name {
		case "created":
			m.Page = 2
		case "list":
			m.JoinPage.AvaibleGames = msg.Data.([]string)
			m.Page = 3
		case "join":
			Data := msg.Data.(map[string]interface{})
			m.Opponent = Data["name"].(string)
			m.OpponentConn = Data["connector"].(*connector.Connector)

			//TODO: confirmation Page if yes send hello message else send sorry

			//1. send a hello
			m.OpponentConn.SendMsg(connector.Msg{Name: "hello"})
			oppmsg, _ := m.OpponentConn.GetMsg()
			//2. wait for hello
			if oppmsg.Name == "hello" {
				m.OpponentStatus = "Connected"
				m.Page = 2
			}
			//3. send first?
			m.OpponentConn.SendMsg(connector.Msg{Name: "first?"})
			//4. if yes game.player = O
			oppmsg, _ = m.OpponentConn.GetMsg()
			if oppmsg.Name == "yes" {
				m.Game.Player = "O"
			} else {
				m.Game.Player = "X"
			}
			return m, tea.Batch(ListentoOpponent(m), ListentoServer(m))
		case "connect":
			m.OpponentConn = msg.Data.(*connector.Connector)

			//4. send first move
			//1. wait for hello
			oppmsg, _ := m.OpponentConn.GetMsg()
			if oppmsg.Name != "hello" {
				m.OpponentStatus = oppmsg.Name
				return m, ListentoServer(m)
			}
			//2. send hello
			m.OpponentConn.SendMsg(connector.Msg{Name: "hello"})
			m.Page = 2
			m.OpponentStatus = "Connected"

			//3. wait for first?
			oppmsg, _ = m.OpponentConn.GetMsg()
			if oppmsg.Name != "first?" {
				m.OpponentStatus = oppmsg.Name
				return m, ListentoServer(m)
			}
			//3. send random() yes/no yes game.player = X for tictactoe
			if rand.Intn(2) == 1 {
				m.OpponentConn.SendMsg(connector.Msg{Name: "yes"})
				m.Game.Player = "X"
			} else {
				m.OpponentConn.SendMsg(connector.Msg{Name: "no"})
				m.Game.Player = "O"
			}
			return m, tea.Batch(ListentoServer(m), ListentoOpponent(m))
		case "left":
			m.OpponentStatus = "Disconnected ..."
		case "error":
			//TODO: error can occur while in any of the actions create, join, list
			//the processing of the request is happening while pages in on loading
			// so if error occurs we should show the error msg from msg.data and start a ticker of 4 5 sec
			// after 4 5 sec the page will be changed to previous page or the home page
		case "exit":
			return m, tea.Quit

		case "move":
			move := msg.Data.([2]int)
			m.Game.Board[move[0]][move[1]] = m.Game.CurrentPlayer
            m.Game.CurrentPlayer = m.Game.Player
			return m, ListentoOpponent(m)

		}
		return m, ListentoServer(m)
	}
	return m, nil
}

func (m Model) View() string {
	s := "\n\n\t\tTick Tac Toe"
	switch m.Page {
	// Page 1 will contain Start screen
	case 1:
		s += "\n\n"
		s += "\t\tO    Join   Game\n"
		s += "\t\tN    Create Game\n"
		s += "\t\tQ    Quit\n"
		s += "\n\n\t\tQ: quit  r:restart\n"
		return s
		//Page 2 will content the Game
	case 2:
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
	case 3:
		// s += " - Join Game \n Input Field : " + m.JoinPage.Input + "\n"
		s += "\n"
		log.Println(len(m.JoinPage.AvaibleGames))
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
	case 4:
		s += "\n\n\n\n"
		s += "\t\tLoading ..."
		s += "\n\n\n\n"
		s += "\t\tQ Quit\n"
	}
	return s
}
