package ssh

import (
	"context"
	"errors"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/log"
	"github.com/charmbracelet/ssh"
	"github.com/charmbracelet/wish"
	"github.com/charmbracelet/wish/activeterm"
	"github.com/charmbracelet/wish/bubbletea"
	"github.com/charmbracelet/wish/logging"

	"github.com/metallust/chessh/internal/connector"
	"github.com/metallust/chessh/internal/tictactoe"
)

const (
	host = "localhost"
	port = "23234"
)

type User struct {
	connection *connector.Connector
	stage      string
	Opponent   string
}

var Users map[string]User = make(map[string]User)

func CreateGame(user string) {
	//add user to ready
	//BUG: rewrite this
	u := Users[user]
	u.stage = "ready"
	Users[user] = u
	//send ok
	Users[user].connection.SendMsg("created", nil)
	log.Info(user, "user ready...", Users)
}


func JoinGame(user string, opponent string) error {
	//if  not in ready return error
	if Users[user].stage != "lobby" {
		return errors.New("You are not allow your stage should be Inital")
	}
	if Users[opponent].stage != "ready" {
		return errors.New("Opponent is no longer avaiable ..(maybe be entered a different game or went offline)")
	}

	//TODO: Ask user for confirmation
	// first way: create a channel and send request to the opponent so when opponent want to say yes just pass yes in that channel
	// second way: create a request pool when request is send and a reference of that request in the pool and wait when the opponent sends ok in the gofunc lister complete the request from the pool and resume the execution

	//send opponent join message with name of user
	Users[opponent].connection.SendMsg("join", user)

	u := Users[user]
	u.Opponent = opponent
	u.stage = "ingame"
	Users[user] = u

	o := Users[opponent]
	o.Opponent = user
	o.stage = "ingame"
	Users[user] = o

	// randomly decide who is going first
	if rand.Intn(2) == 0 {
        log.Info("opponent is first")
		Users[user].connection.SendMsg("first", false)
		Users[opponent].connection.SendMsg("first", true)
	} else {
        log.Info("user is first")
		Users[user].connection.SendMsg("first", true)
		Users[opponent].connection.SendMsg("first", false)
	}
	return nil
}

func ListGames(user string) {
	data := make([]string, 0)
	for k, v := range Users {
		if k != user && v.stage == "ready" {
			data = append(data, k)
		}
	}
	Users[user].connection.SendMsg("list", data)
}

func ExitGame(user string) {
    userConn := Users[user].connection
    if userConn != nil {
        userConn.Close()
    }
    // if game is in progress
    if Users[user].Opponent != "" {
        // send opponent the opponent abort msg
        Users[Users[user].Opponent].connection.SendMsg("abort", nil)
        // send opponent to the lobby
        o := Users[Users[user].Opponent]
        o.Opponent = ""
        o.stage = "lobby"
        Users[Users[user].Opponent] = o
    }
    // remove user from Users list
    delete(Users, user)
    log.Info("User removed", "user", user, "Users", Users)
}
func Move(user string, move [2]int) error {
	opponentStr := Users[user].Opponent
	if opponentStr == "" {
		return errors.New("No opponent found !!!")
	}
	Users[opponentStr].connection.SendMsg("move", move)
	return nil
}

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	//generate random blob
	blob := s.Context().SessionID()
	user := s.User() + blob[:5]

	// NOTE : Actual user is create here
	c := connector.NewConnector()
	cpair := connector.CreateConnectorPair(c)
	Users[user] = User{
		connection: c,
		stage:      "lobby",
	}
	go func() {
		for {
			clientMsg, more := Users[user].connection.GetMsg()
			if more == false {
				log.Info("Channel closed stopping go routing", "user", user)
				return
			}
			switch clientMsg.Name {
			//TODO: add error handling if the function return any error forward that to client
			case "create":
				log.Info("Creating game for ", "user", user)
				CreateGame(user)
			case "list":
				log.Info("Listing avaiable games for", "user", user)
				ListGames(user)
			case "join":
				opponent := clientMsg.Data.(string)
				//ask apponent for confirmation
				log.Info("joining ...", "User", opponent, "User ", user)
				JoinGame(user, opponent)
			case "move":
				move := clientMsg.Data.([2]int)
				log.Info("Moved", "user", user, "move", move)
				Move(user, move)
			case "exit":
				log.Info("Exiting ... ", "User ", user)
                break
			}
		}
	}()

	log.Info(Users)
	m := tictactoe.InitialModel(user, cpair)
	return m, []tea.ProgramOption{tea.WithAltScreen()}
}

func StartServer() {
	s, err := wish.NewServer(wish.WithAddress(net.JoinHostPort(host, port)), wish.WithHostKeyPath(".ssh/id_ed25519"), wish.WithMiddleware(
		bubbletea.Middleware(teaHandler),
		activeterm.Middleware(),
		logging.Middleware(),
		func(next ssh.Handler) ssh.Handler {
			return func(s ssh.Session) {
				next(s)
				user := s.User() + s.Context().SessionID()[:5]
				log.Info("User disconnected", "user", user)
				ExitGame(user)
			}
		},
	))

	if err != nil {
		log.Error("Could not start server", "error", err)
	}

	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)
	log.Info("Starting SSH server", "host", host, "port", port)
	go func() {
		if err = s.ListenAndServe(); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
			log.Error("Could not start server", "error", err)
			done <- nil
		}
	}()

	<-done
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}
