package game

import (
	"errors"
	"net"

	log "github.com/inconshreveable/log15"
	"fmt"
)

type GamePlan struct {
	N   int
	Map [][]int
}

type Client struct {
	Connection net.Conn
}

type Game struct {
	id string
	ready bool
	player_on_move int
	players []Client
	state 	string
	plan  	GamePlan
}

func NewGame(id string, n int) *Game {
	log.Debug("Creating new game")

	// initialize plan
	plan := GamePlan{
		N: n,
	}

	plan.Map = make([][]int, plan.N)
	for i := range plan.Map {
		plan.Map[i] = make([]int, plan.N)
		for j := range plan.Map[i] {
			plan.Map[i][j] = -1
		}
	}

	// initialize and return game
	return &Game{
		id: id,
		ready: false,
		player_on_move: 0,
		players: make([]Client, 2),
		state: "empty",
		plan: plan,
	}
}

func (this *Game) AddClient(conn net.Conn) (int, error) {
	client := Client{
		Connection: conn,
	}

	switch this.state {
	case "empty":
		this.players[0] = client
		this.state = "awaitingoponent"
		return 0, nil
	case "awaitingoponent":
		this.players[1] = client
		this.state = "play"
		this.ready = true
		return 1, nil
	default:
		return -1, errors.New("Too many clients")
	}
}

// get current game ID
func (this *Game) ID() string {
	return this.id
}

// get current game state
func (this *Game) State() string {
	return this.state
}

// print current game plan
func (this *Game) Print() {
	for i := 0; i < this.plan.N; i++ {
		for j := 0; j < this.plan.N; j++ {
			fmt.Printf("|%d", this.plan.Map[i][j])
		}
		fmt.Println("|")
	}
}

// returns player on move
func (this *Game) OnMove() int {
	if !this.Ready() {
		return -1
	}

	return this.player_on_move
}

func (this *Game) Ready() bool {
	return this.ready
}