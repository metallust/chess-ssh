package ssh

import (
	"context"
	"errors"
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

func ExitGame(user string) {
	// if game is in progress
	if Users[user].Opponent != "" {
		// send opponent the opponent abort msg
		Users[Users[user].Opponent].connection.SendMsg("abort", nil)
		// opponent.Opponet = nil
		// close opponent channel
	}
	// close server channel
	Users[user].connection.Close()
	// remove user from Users list
	delete(Users, user)
	//
}

func JoinGame(user string, opponent string) {
	//if  not in ready return error
	if Users[user].stage != "initial" {
		Users[user].connection.SendMsg("error", "You are not allow your stage should be Inital")
		return
	}
	if Users[opponent].stage != "ready" {
		Users[user].connection.SendMsg("error", "Opponent is no longer avaiable ..(maybe be entered a different game or went offline)")
		return
	}

	//NOTE: opponent pair connector are created and one is given to user and one is given to opponent
	usrconn := connector.NewConnector()
	oppconn := connector.CreateConnectorPair(usrconn)

	//send the message to the opponent that this user wants to connect with channel
	Users[opponent].connection.SendMsg("join", map[string]interface{}{"name": user, "connector": oppconn})
	//if the opponent accepts(ok) send the user the ok
	Users[user].connection.SendMsg("connect", usrconn)
	//BUG: rewrite this
	//add user to opponent.Opponent
	//add oppenent to user.Opponent
	u := Users[opponent]
	u.Opponent = user
	Users[opponent] = u

	u = Users[user]
	u.Opponent = opponent
	Users[opponent] = u
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

func teaHandler(s ssh.Session) (tea.Model, []tea.ProgramOption) {
	//generate random blob
	blob := s.Context().SessionID()
	user := s.User() + blob[:5]

	// NOTE : Actual user is create here
	c := connector.NewConnector()
	cpair := connector.CreateConnectorPair(c)
	Users[user] = User{
		connection: c,
		stage:      "initial",
	}
	go func() {
		for {
			log.Info("Loopping ...")
			clientMsg, more := Users[user].connection.GetMsg()
			if more == false {
				log.Info("Channel closed", user)
				return
			}
			switch clientMsg.Name {
			case "create":
				log.Info("Creating game for ", user)
				CreateGame(user)
			case "list":
				log.Info("Listing avaiable games for", user)
				ListGames(user)
			case "join":
				opponent := clientMsg.Data.(string)
				log.Info("joining ...", "User", opponent, "User ", user)
				JoinGame(user, opponent)
			case "exit":
				log.Info("Exiting ... ", "User ", user)
				ExitGame(user)
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
				log.Info("User disconnected", "user", user, "Exiting from Game")
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

	//close the channels in users
	log.Info("Stopping SSH server")
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer func() { cancel() }()
	if err := s.Shutdown(ctx); err != nil && !errors.Is(err, ssh.ErrServerClosed) {
		log.Error("Could not stop server", "error", err)
	}
}
