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
			clientMsg, more := c.GetMsg()
			if more == false {
				log.Info("Channel closed stopping go routing", "user", user)
				return
			}
			switch clientMsg.Name {
			//TODO: add error handling if the function return any error forward that to client
			case "create":
				log.Info("Creating game for ", "user", user)
				CreateGame(user, clientMsg)
			case "list":
				log.Info("Listing avaiable games for", "user", user)
				ListGames(user, clientMsg)
			case "join":
				opponent := clientMsg.Data.(string)
				//ask apponent for confirmation
				log.Info("joining ...", "User", opponent, "User ", user)
				JoinGame(user, clientMsg)
			case "move":
				move := clientMsg.Data.([2]int)
				log.Info("moved", "user", user, "move", move)
				Move(user, clientMsg)
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

func randomPlayer() (string, string) {
	// randomly decide who is going first
	if randbool := rand.Intn(2) == 0; randbool {
		return "first", "second"
	}
	return "second", "first"
}
