package game

import (
	"errors"
	"fmt"
	"net"
	"strconv"

	log "github.com/inconshreveable/log15"
)

type GamePlan struct {
	n               int
	plan            [][]int
	freeFieldsCount int
}

type Client struct {
	PlayerName string
	Connection net.Conn
}

type Game struct {
	id string
	ready bool
	player_on_move int
	players []*Client
	state 	string
	plan  	GamePlan
}

func NewGame(id string, n int) *Game {
	log.Debug("Creating new game")

	// initialize plan
	plan := GamePlan{
		n:               n,
		freeFieldsCount: n*n,
	}

	plan.plan = make([][]int, plan.n)
	for i := range plan.plan {
		plan.plan[i] = make([]int, plan.n)
		for j := range plan.plan[i] {
			plan.plan[i][j] = -1
		}
	}

	// initialize and return game
	return &Game{
		id: id,
		ready: false,
		player_on_move: 0,
		players: make([]*Client, 2),
		state: "empty",
		plan: plan,
	}
}

func (this *Game) AddClient(conn net.Conn, name string) (int, error) {
	client := &Client{
		Connection: conn,
		PlayerName: name,
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
		log.Debug("SENDING MOVE -1 -1")
		this.players[0].Connection.Write([]byte("MOVE -1 -1\n"))
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
	for i := 0; i < this.plan.n; i++ {
		for j := 0; j < this.plan.n; j++ {
			var c string
			if this.plan.plan[i][j] == -1 {
				c = " "
			} else if this.plan.plan[i][j] == 0 {
				c = "X"
			} else if this.plan.plan[i][j] == 1 {
				c = "O"
			}
			fmt.Printf("|%s", c)
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

func (this *Game) Move(player_id int, x int, y int) error {
	if x > this.plan.n - 1 || y > this.plan.n- 1 ||
		x < 0 || y < 0 || this.plan.plan[y][x] != -1 {
			return errors.New("Invalid move")
	}
	log.Debug("Player move", "x", x, "y", y, "player_id", player_id, "player_name", this.players[player_id].PlayerName)
	this.plan.plan[y][x] = player_id
	this.plan.freeFieldsCount--
	if winner, end := this.CheckWin(player_id, x, y); end {
		this.GameEnd(winner)
	} else {
		this.players[len(this.players) - player_id - 1].Connection.Write([]byte("MOVE " +
			strconv.FormatUint(uint64(x), 10) + " " +
			strconv.FormatUint(uint64(y), 10) + "\n"))
	}

	this.player_on_move = len(this.players) - player_id - 1
	this.Print()

	return nil
}

func (this *Game) GameEnd(winner int) {
	for _, player := range this.players {
		if player == nil {
			continue
		}
		player.Connection.Write([]byte("GAMEEND " + strconv.FormatInt(int64(winner), 10) + "\n"))
		player.Connection.Close()
	}
}

func (this *Game) CheckWin(player_id int, last_x int, last_y int) (int, bool) {
	var n int = 5
	if this.plan.n < 5 {
		n = this.plan.n
	}
	var count int

	count = 0
	for x := last_x; x >= 0; x-- {
		if this.plan.plan[last_y][x] == player_id {
			count++
		} else {
			break
		}
	}
	for x := last_x + 1; x < this.plan.n; x++ {
		if this.plan.plan[last_y][x] == player_id {
			count++
		} else {
			break
		}
	}
	if count >= n {
		log.Debug("GAMEEND", "player_id", player_id)
		return player_id, true
	}

	count = 0
	for y := last_y; y >= 0; y-- {
		if this.plan.plan[y][last_x] == player_id {
			count++
		} else {
			break
		}
	}
	for y := last_y + 1; y < this.plan.n; y++ {
		if this.plan.plan[y][last_x] == player_id {
			count++
		} else {
			break
		}
	}
	if count >= n {
		log.Debug("GAMEEND", "player_id", player_id)
		return player_id, true
	}

	count = 0
	for x, y := last_x, last_y; x >= 0 && y >= 0; x, y = x - 1, y - 1 {
		log.Debug("Checking diagonal", "x", x, "y", y, "N", this.plan.n)
		if this.plan.plan[y][x] == player_id {
			count++
		} else {
			break
		}
	}
	for x, y := last_x + 1, last_y + 1; x < this.plan.n && y < this.plan.n; x, y = x + 1, y + 1 {
		log.Debug("Checking diagonal", "x", x, "y", y, "N", this.plan.n)
		if this.plan.plan[y][x] == player_id {
			count++
		} else {
			break
		}
	}
	if count >= n {
		log.Debug("GAMEEND", "player_id", player_id)
		return player_id, true
	}

	count = 0
	for x, y := last_x, last_y; x < this.plan.n && y >= 0; x, y = x + 1, y - 1 {
		log.Debug("Checking diagonal - the other one", "x", x, "y", y, "N", this.plan.n)
		if this.plan.plan[y][x] == player_id {
			count++
		} else {
			break
		}
	}
	for x, y := last_x - 1, last_y + 1; x >= 0 && y < this.plan.n; x, y = x - 1, y + 1 {
		log.Debug("Checking diagonal - the other one", "x", x, "y", y, "N", this.plan.n)
		if this.plan.plan[y][x] == player_id {
			count++
		} else {
			break
		}
	}
	if count >= n {
		log.Debug("GAMEEND", "player_id", player_id)
		return player_id, true
	}


	if(this.plan.freeFieldsCount == 0) {
		return -1, true
	} else {
		return -2, false
	}
}