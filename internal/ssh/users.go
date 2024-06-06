package ssh

import (
	"github.com/charmbracelet/log"
	"github.com/metallust/chessh/internal/connector"
    "sync"
)

type User struct {
	connection *connector.Connector
	stage      string
	Opponent   string
}

var (
    Users map[string]User = make(map[string]User)
    UserMut sync.Mutex
)

func CreateGame(user string, msg connector.Msg) {
	//check if the user is in the lobby
    UserMut.Lock()
	if Users[user].stage != "lobby" {
		msg.Reply("error", "My guy you are not in the lobby - (can be mutex issue)", false)
	}

	//add user to ready
	//BUG: rewrite this
	u := Users[user]
	u.stage = "ready"
	Users[user] = u
    UserMut.Unlock()
	msg.Reply("ok", nil, false)
}

func JoinGame(user string, msg connector.Msg) {
    UserMut.Lock()
	opponent := msg.Data.(string)
	//if  not in ready return error
	if Users[user].stage != "lobby" {
		msg.Reply("error", "You are not allow your stage should be Inital", false)
		return
	}
	if Users[opponent].stage != "ready" {
		msg.Reply("error", "Opponent is no longer avaiable ..(maybe be entered a different game or went offline)", false)
		return
	}

	//TODO: Ask user for confirmation
	// first way: create a channel and send request to the opponent so when opponent want to say yes just pass yes in that channel
	// second way: create a request pool when request is send and a reference of that request in the pool and wait when the opponent sends ok in the gofunc lister complete the request from the pool and resume the execution

	//decide turn
	playerA, playerB := randomPlayer()
	oppreqdata := map[string]string{
		"opponent": user,
		"turn":     playerA,
	}

	//send opponent join message with name of user
	oppreply := Users[opponent].connection.SendMsg("join", oppreqdata, true)
	oppreplymsg := <-oppreply
	if oppreplymsg.Name != "ok" {
		msg.Reply("error : opponent replied"+oppreplymsg.Name, oppreplymsg.Data, false)
		return
	}

	//replying user
	msg.Reply("ok", playerB, false)

	u := Users[user]
	u.Opponent = opponent
	u.stage = "ingame"
	Users[user] = u

	o := Users[opponent]
	o.Opponent = user
	o.stage = "ingame"
	Users[opponent] = o

    UserMut.Unlock()
	log.Info("Game started", "user", user, "opponent", opponent)
}

func ListGames(user string, msg connector.Msg) {
	data := make([]string, 0)
    UserMut.Lock()
	for k, v := range Users {
		if k != user && v.stage == "ready" {
			data = append(data, k)
		}
	}
    UserMut.Unlock()
	msg.Reply("list", data, false)
}

func ExitGame(user string) {

    UserMut.Lock()
	userConn := Users[user].connection
	if userConn != nil {
		userConn.Close()
	}
	// if game is in progress
	opponent := Users[user].Opponent
	if opponent != "" {
		// send opponent the opponent abort msg
		// send opponent to the lobby
		o := Users[opponent]
		o.Opponent = ""
		o.stage = "lobby"
		Users[Users[user].Opponent] = o

		Users[opponent].connection.SendMsg("abort", nil, false)
	}
	// remove user from Users list
	delete(Users, user)
    UserMut.Unlock()
	log.Info("User removed", "user", user, "Users", Users)
}

func Move(user string, msg connector.Msg) {
    UserMut.Lock()
	move := msg.Data.([2]int)
	opponentStr := Users[user].Opponent
	if opponentStr == "" {
		msg.Reply("error", "Opponent not found", false)
		return
	}
	replychan := Users[opponentStr].connection.SendMsg("move", move, true)
	if reply := <-replychan; reply.Name != "ok" {
		msg.Reply("error", "Opponent is not accepting your move", false)
		return
	}
    UserMut.Unlock()
	msg.Reply("ok", nil, false)
}

