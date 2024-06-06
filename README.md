# Chess over SSH

## Tech Used
1. Golang
2. Bubble Tea
3. Wish

## How to run
To start the SSH server that will server a bubble tea chess application run 
```shell
make setup
make run
```
SSH will start and to connect to is server
1. Open a new terminal
2. ssh -p 23234 localhost
3. Start 2 instance for testing


## Things to add:
1. ~~Add mutex for Users[]~~
2. abort msg handler 
3. Figure out a better way to log bubble tea app (maybe in a file or on the stdout with log of server)
4. Add style
5. add generics
6. Organize file pagewise
7. use list bubble tea component in joinpage

<!-- Think to learn
1. Project Structure
2. Goroutines
3. Channels
4. Pointers
6. Context
5. Mutex  -->



5 MAY 2024
