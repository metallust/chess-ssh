package main

import (
	"github.com/metallust/chessh/pkg/tictactoe"
	"github.com/metallust/sshGameClient/server"
)

func main() {
    server.StartServer(tictactoe.InitialModel)
}
