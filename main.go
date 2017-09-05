package main

import (
	"bufio"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	log "github.com/inconshreveable/log15"

	"git.tumeo.eu/lstme/tictactoe-server/game"
)

var n, _ = strconv.ParseInt(os.Getenv("GAME_N"), 10, 32)
var N = int(n)

type GameArray struct {
	sync.Mutex
	Games map[string]*game.Game
}

var games_store = GameArray{
	Games: map[string]*game.Game{},
}

func main() {
	if N < 3 {
		log.Warn("N < 3 => N=3")
		N = 3
	}

	ln, err := net.Listen("tcp", ":32768")
	if err != nil {
		log.Error("Cannot bind to port")
		panic("Cannot bind to port")
	}

	for {
		conn, err := ln.Accept()
		if err != nil {
		}
		go handleNewConnection(conn)
	}
}



func handleNewConnection(conn net.Conn) {
	log.Info("New connection")
	//conn.SetDeadline(time.Now().Add(time.Duration(30) * time.Second))
	// Close the connection when you're done with it.
	defer conn.Close()

	var my_game *game.Game
	var my_id int = -1
	// Make a buffer to hold incoming data.
	bufReader := bufio.NewReader(conn)

	// assign this player to a new game
	for {
		// Read commands delimited by newline
		buf, err := bufReader.ReadBytes('\n')
		if err != nil {
			log.Error("Error reading:", "error", err.Error())
			//break;
			return
		}

		// parse `HELO id` command
		matches := strings.Split(string(buf), " ")
		if len(matches) < 3 {
			log.Debug("No HELO command was sent", "data", matches)
			conn.Write([]byte("NOHELO\n"))
			continue
			//return
		}
		game_id := matches[1]

		games_store.Lock()

		log.Debug("Connecting to game", "game_id", game_id, "len(games_store.Games)", len(games_store.Games))
		var ok bool
		if my_game, ok = games_store.Games[game_id]; !ok {
			my_game = game.NewGame(game_id, N)
			games_store.Games[game_id] = my_game
		}

		// add new client/player to this game
		if my_id, err = my_game.AddClient(conn, strings.Trim(matches[2], "\n")); err != nil {
			log.Warn("GAMEFULL")
			conn.Write([]byte("GAMEFULL\n"))
			my_game = nil
			continue
			//return
		}
		log.Debug("Client connected", "game_id", game_id, "data", matches, "player_id", my_id)

		conn.Write([]byte("OK " +
			game_id + " " +
			strconv.FormatInt(int64(N), 10) + " " +
			strconv.FormatInt(int64(my_id), 10) + "\n"))
		my_game.Print()
		games_store.Unlock()
		break
	}

	for {
		buf, err := bufReader.ReadBytes('\n')
		if err != nil {
			games_store.Lock()
			log.Error("Error reading:", "error", err.Error(), "game_id", my_game.ID())
			delete(games_store.Games, my_game.ID())
			my_game.GameEnd(-2)
			my_game = nil
			games_store.Unlock()
			//break;
			return
		}

		if my_game.OnMove() != my_id || !my_game.Ready() {
			conn.Write([]byte("OUTOFORDER\n"))
			log.Warn("User move out of order", "my_id", my_id, "on_move", my_game.OnMove())
			continue
		}

		matches := strings.Split(strings.Trim(string(buf), " \r\n"), " ")
		switch matches[0] {
		case "MOVE":
			if len(matches) != 3 {
				log.Error("Invalid move command")
				conn.Write([]byte("ERROR\n"))
				continue
			}
			x, err := strconv.Atoi(matches[1])
			y, err := strconv.Atoi(matches[2])
			if err != nil {
				log.Error("Invalid x y", "x", matches[1], "y", matches[2])
				conn.Write([]byte("ERROR\n"))
				continue
			}

			if err := my_game.Move(my_id, x, y); err != nil {
				log.Debug("CANNOT", "player_id", my_id)
				conn.Write([]byte("CANNOT\n"))
				continue
			}
		default:
			conn.Write([]byte("ERROR\n"))
		}
	}
}
