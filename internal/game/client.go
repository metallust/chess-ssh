package game

import (
	"errors"
	"log"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/metallust/chessh/internal/connector"
)

type GameClientMsg struct{
    Msg string
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
        log.Println("Bubble tea game client", msg, more)
		if !more {
			log.Println("Bubbletea Application: Server disconnected ...")
            return GameClientMsg{Msg: "exit"}
		}
		switch msg.Name {
		case "join":
			first, opponentname, err := gc.AcceptRequest(msg.Data.(string))
            if err != nil {
                return GameClientMsg{Msg: "error", Data: err.Error()}
            }
            return GameClientMsg{Msg: "joined", Data: map[string]interface{}{"first": first, "opponent": opponentname}}
		case "error":
			//TODO: error can occur while in any of the actions create, join, list
			//the processing of the request is happening while pages in on loading
			// so if error occurs we should show the error msg from msg.data and start a ticker of 4 5 sec
			// after 4 5 sec the page will be changed to previous page or the home page
            return GameClientMsg{Msg: "error", Data: msg.Data}
		case "exit":
            return GameClientMsg{Msg: "exit"}
		case "move":
            return GameClientMsg{Msg: "move", Data: msg.Data}
		}
        return GameClientMsg{Msg: "different"}
	}
}

func (gc *GameClient) List() ([]string, error) {
    log.Println("Bubbletea Application: Fetching List ...")
	gc.serverconnector.SendMsg("list", nil)
	msg, more := gc.serverconnector.GetMsg()
	if !more {
		//add handler to server close connection
	}
	if msg.Name == "error" {
		//handle that error
		return nil, errors.New(msg.Data.(string))
	}
	return msg.Data.([]string), nil
}

func (gc *GameClient) Create() error {
	log.Println("Bubbletea Application: Creating...")
	gc.serverconnector.SendMsg("create", nil)
	msg, more := gc.serverconnector.GetMsg()
	if !more {
		//add handler to server close connection
	}
	if msg.Name == "error" {
		//handle that error
		return errors.New(msg.Data.(string))
	}
	return nil
}

func (gc *GameClient) Move(move [2]int) error {
	log.Println("Bubbletea Application: Making Move ...")
	gc.serverconnector.SendMsg("move", move)
	msg, more := gc.serverconnector.GetMsg()
	if !more {
		//add handler to server close connection
	}
	if msg.Name == "error" {
		//handle that error
		return errors.New(msg.Data.(string))
	}
	return nil
}

func (gc *GameClient) Join(opponent string) (bool, error) {
	log.Println("Bubbletea Application: Joining ...")

	gc.serverconnector.SendMsg("join", opponent)
	msg, more := gc.serverconnector.GetMsg()
	if !more {
		//add handler to server close connection
	}
	if msg.Name == "error" {
		//handle that error
		return false, errors.New(msg.Data.(string))
	}
	if msg.Name != "first" {
		return false, errors.New("something went wrong: wrong protocol")
	}
    log.Println("Bubbletea Application: Joined ...")
	return msg.Data.(bool), nil
}

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
