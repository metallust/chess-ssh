package main

import (
	"github.com/metallust/chessssh/pkg/tictactoe"
	"github.com/metallust/dosssh/server"
)

func main() {
    server.StartServer(tictactoe.InitialModel)
}
