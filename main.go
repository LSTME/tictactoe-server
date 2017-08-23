package main

import (
	"net"
	"regexp"
	"strconv"
	"bufio"

	log "github.com/inconshreveable/log15"

	"git.tumeo.eu/lstme/tictactoe-server/game"
	"time"
)

const N = 3

var games map[string]*game.Game = map[string]*game.Game{}

func main() {
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
	conn.SetDeadline(time.Now().Add(time.Duration(30) * time.Second))
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
			break;
		}

		// parse `HELO id` command
		r := regexp.MustCompile(`HELO (?P<id>.+)`)
		matches := r.FindStringSubmatch(string(buf))
		if len(matches) < 2 {
			conn.Write([]byte("NOHELO\n"))
			continue
			//return
		}
		game_id := matches[1]

		var ok bool
		if my_game, ok = games[game_id]; !ok {
			my_game = game.NewGame(game_id, N)
			games[game_id] = my_game
		}

		// add new client/player to this game
		if my_id, err = my_game.AddClient(conn); err != nil {
			conn.Write([]byte("GAMEFULL\n"))
			continue
			//return
		}

		conn.Write([]byte("OK " +
							game_id + " " +
							strconv.FormatUint(uint64(N), 10) + " " +
							strconv.FormatUint(uint64(my_id), 10) + "\n"))
		my_game.Print()
		break;
	}

	for {
		buf, err := bufReader.ReadBytes('\n')
		if err != nil {
			log.Error("Error reading:", "error", err.Error())
			break;
		}

		if my_game.OnMove() != my_id || !my_game.Ready() {
			conn.Write([]byte("OUTOFORDER\n"))
			log.Warn("User move out of order", "my_id", my_id, "on_move", my_game.OnMove())
		}

		var _ = buf
	}
}