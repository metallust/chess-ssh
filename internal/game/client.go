package game

import (
	"errors"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/metallust/chessh/internal/connector"
)

type GameClientMsg struct {
	Msg  string
	Data interface{}
}

type DoneMsg struct {
	Msg  string
	Data interface{}
}
type GameClient struct {
	serverconnector *connector.Connector
}

func NewGameClient(c *connector.Connector) *GameClient {
	return &GameClient{
		serverconnector: c,
	}
}

func (gc *GameClient) ListenServer() tea.Cmd {
	return func() tea.Msg {
		msg, more := gc.serverconnector.GetMsg()
		log.Println("Bubble tea Lister server -->", msg, more)
		if !more {
			log.Println("Bubbletea Application: Server disconnected ...")
			return GameClientMsg{Msg: "exit"}
		}
		switch msg.Name {
		case "join":
			// join msg ask for confirmation
			// send ok msg through the replyfunc for confirmation

			//TODO: prompt the user by emitting a yes/no message
			//let the user decide if he wants to play with the opponent
			//if yes then he has to run acceptJoinRequest function else declineJoinRequest func
			msg.Reply("ok", nil, false)

			//have to turn this into custom msg where user can set the emmit msg
			return GameClientMsg{Msg: "join", Data: msg.Data}
		case "exit":
			return GameClientMsg{Msg: "exit"}
		case "move":
            msg.Reply("ok", nil, false)
			return GameClientMsg{Msg: "move", Data: msg.Data}
		case "error":
			//TODO: error can occur while in any of the actions create, join, list
			//the processing of the request is happening while pages in on loading
			// so if error occurs we should show the error msg from msg.data and start a ticker of 4 5 sec
			// after 4 5 sec the page will be changed to previous page or the home page
			return GameClientMsg{Msg: "error", Data: msg.Data}
		}
		return GameClientMsg{Msg: "unknown"}
	}
}

// this function return as tea msg which is gameclientmsg which contain name and data
// name is set to the donemsg given by the caller
// working: this will send "list" message to the server and wait for the reply
// the server if successfull will return "ok"
// if there is any error server will return "error" and the data will be the error message which will be forwarded to
// the caller will msg name set to "error" and data set to the error message
func (gc *GameClient) List(doneMsg string) tea.Cmd {
	return func() tea.Msg {
		log.Println("Bubbletea Application: Fetching List ...")
		replychan := gc.serverconnector.SendMsg("list", nil, true)
		msg := <-replychan
		if msg.Name == "error" {
			//handle that error
            log.Fatal("Error in list ... Here is the msg : ", msg)
			return DoneMsg{Msg: "error", Data: msg.Data}
		}
		return DoneMsg{Msg: doneMsg, Data: msg.Data.([]string)}
	}
}

// this function return as tea msg which is gameclientmsg which contain name and data
// name is set to the donemsg given byc the caller
// working: this will send "create" message to the server and wait for the reply
// the server if successfull will return "ok"
// if there is any error server will return "error" and the data will be the error message which will be forwarded to
// the caller will msg name set to "error" and data set to the error message
func (gc *GameClient) Create(doneMsg string) tea.Cmd {
	return func() tea.Msg {
		log.Println("Bubbletea Application: Creating...")
		replychan := gc.serverconnector.SendMsg("create", nil, true)
		msg, _ := <-replychan
        if msg.Name != "ok" {
            log.Fatal("Error in Create ... Here is the msg : ", msg)
            return DoneMsg{Msg: "error", Data: msg.Data}
        }
		return DoneMsg{Msg: doneMsg}
	}
}

// this function return as tea msg which is gameclientmsg which contain name and data
// name is set to the donemsg given by the caller
// working: this will send "move" message and move array to the server and wait for the reply
// the server if successfull will return "ok"
// if there is any error server will return "error" and the data will be the error message which will be forwarded to
// the caller will msg name set to "error" and data set to the error message
func (gc *GameClient) Move(move [2]int, doneMsg string) tea.Cmd {
	return func() tea.Msg {
		log.Println("Bubbletea Application: Making Move ...")
		replychan := gc.serverconnector.SendMsg("move", move, true)
		msg, _ := <-replychan
		if msg.Name != "ok" {
			//handle that error
            log.Fatal("Error in Move ... Here is the msg : ", msg)
			return DoneMsg{Msg: "error", Data: msg.Data}
		}
		return DoneMsg{Msg: doneMsg}
	}
}

// this function return as tea msg which is gameclientmsg which contain name and data
// name is set to the donemsg given by the caller
// working: this will send "join" message and opponent name string to the server and wait for the reply
// the server if successfull will return "ok" and the data will be either "first" or "second"
// if there is any error server will return "error" and the data will be the error message which will be forwarded to
// the caller will msg name set to "error" and data set to the error message
func (gc *GameClient) Join(opponent, doneMsg string) tea.Cmd {
	return func() tea.Msg {
		log.Println("Bubbletea Application: Joining ...")
		replychan := gc.serverconnector.SendMsg("join", opponent, true)
		msg, _ := <-replychan
		log.Println("Bubbletea Application: Join reply ...", msg)
		if msg.Name != "ok" || (msg.Data != "first" && msg.Data != "second") {
			//handle that error
            log.Fatal("Error in Join ... Here is the msg : ", msg)
			return DoneMsg{Msg: "error", Data: msg.Data}
		}
		return DoneMsg{Msg: doneMsg, Data: msg.Data}
	}
}

// FIX:
// Internal function to accept the join request from the opponent
func (gc *GameClient) AcceptRequest(opponent string) (bool, string, error) {
	log.Println("Bubbletea Application: Accepting join request...")
	//TODO: confirmation Page if yes send hello message else send sorry

	msg, more := gc.serverconnector.GetMsg()
	if !more {
		//add handler to server close connection
	}
	if msg.Name == "error" {
		//handle that error
		return false, "", errors.New(msg.Data.(string))
	}
	if msg.Name != "first" {
		return false, "", errors.New("something went wrong: wrong protocol")
	}
	return msg.Data.(bool), opponent, nil
}
