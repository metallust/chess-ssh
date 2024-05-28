package main

import (
	"github.com/metallust/chessh/internal/ssh"
)

func main() {
    ssh.StartServer()
	fmt.Println("this is cool")
	// p := tea.NewProgram(game.InitialModel("irfan"))
	// if _, err := p.Run(); err != nil {
	// 	fmt.Println("Error running tea Program ", err)
	// 	os.Exit(1 )
	// }
}
