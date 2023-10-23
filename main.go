package main

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"os"
	"os/signal"
	"strconv"
	"time"

	"github.com/mattn/go-tty"
	"golang.org/x/term"
)

type game struct {
	score int
	snake *snake
	food  position
}

type snake struct {
	body      []position
	direction direction
}

type position [2]int

type direction int

const (
	north direction = iota
	east
	south
	west
)

func init() {
	rand.New(rand.NewSource(time.Now().UnixNano()))
}

func newGame() *game {
	snake := newSnake()

	game := &game{
		score: 0,
		snake: snake,
		food:  randomPosition(),
	}

	go game.listenForKeyPress()

	return game
}

func newSnake() *snake {
	maxX, maxY := getSize()
	pos := position{maxX / 2, maxY / 2}

	return &snake{
		body:      []position{pos},
		direction: north,
	}
}

func main() {

	game := newGame()
	game.prepare()
	for {

		x, y := getSize()

		// calculate new head position
		newHeadPos := game.snake.body[0]

		switch game.snake.direction {
		case north:
			newHeadPos[1]--

		case east:
			newHeadPos[0]++
		case south:
			newHeadPos[1]++
		case west:
			newHeadPos[0]--

		}

		hitWall := newHeadPos[0] < 1 || newHeadPos[1] < 1 || newHeadPos[0] > x ||
			newHeadPos[1] > y
		if hitWall {
			fmt.Println("hit wall")
			game.over()
		}
		for _, pos := range game.snake.body {
			if positionsAreSame(newHeadPos, pos) {
				fmt.Println("in self")
				game.over()
			}
		}

		game.snake.body = append([]position{newHeadPos}, game.snake.body...)

		ateFood := positionsAreSame(game.food, newHeadPos)
		if ateFood {
			game.score++
			game.placeNewFood()
		} else {
			game.snake.body = game.snake.body[:len(game.snake.body)-1]
		}

		game.draw()

	}
}

func (g *game) placeNewFood() {
	for {
		newFoodPosition := randomPosition()

		if positionsAreSame(newFoodPosition, g.food) {
			continue
		}

		for _, pos := range g.snake.body {
			if positionsAreSame(newFoodPosition, pos) {
				continue
			}
		}

		g.food = newFoodPosition

		break
	}
}

func positionsAreSame(a, b position) bool {
	return a[0] == b[0] && a[1] == b[1]
}

func (g *game) prepare() {
	hideCursor()

	// handle CTRL C
	c := make(chan os.Signal, 1)
	signal.Notify(c, os.Interrupt)

	go func() {
		for range c {
			fmt.Println("ctrl-c")
			g.over()
		}
	}()
}

func (g *game) draw() {
	clear()
	maxX, _ := getSize()

	status := "得分: " + strconv.Itoa(g.score)
	statusXPos := maxX/2 - len(status)/2

	moveCursor(position{statusXPos, 0})
	draw(cGreen + status)

	moveCursor(g.food)
	draw(cWhite + "*")

	for i, pos := range g.snake.body {
		moveCursor(pos)

		if i == 0 {
			draw("O")
		} else {
			draw("o")
		}
	}

	render()
	time.Sleep(time.Millisecond * 50)
}

func (g *game) over() {
	clear()
	showCursor()

	moveCursor(position{1, 1})
	draw("game over. score: " + strconv.Itoa(g.score))

	render()

	os.Exit(0)
}

func getSize() (int, int) {
	width, height, err := term.GetSize(int(os.Stdout.Fd()))
	if err != nil {
		panic(err)
	}

	return width, height
}

func randomPosition() position {
	width, height := getSize()
	x := rand.Intn(width) + 1
	y := rand.Intn(height) + 2

	return [2]int{x, y}
}
func (g *game) listenForKeyPress() {
	tty, err := tty.Open()
	if err != nil {
		log.Fatal(err)
	}
	defer tty.Close()

	for {
		char, err := tty.ReadRune()
		if err != nil {
			panic(err)
		}

		// UP, DOWN, RIGHT, LEFT == [A, [B, [C, [D
		// we ignore the escape character [
		switch char {
		case 'A':
			g.snake.direction = north
		case 'B':
			g.snake.direction = south
		case 'C':
			g.snake.direction = east
		case 'D':
			g.snake.direction = west
		}
	}
}

var screen = bufio.NewWriter(os.Stdout)

func hideCursor() {
	fmt.Fprint(screen, "\033[?25l")
}

func showCursor() {
	fmt.Fprint(screen, "\033[?25h")
}

func moveCursor(pos [2]int) {
	fmt.Fprintf(screen, "\033[%d;%dH", pos[1], pos[0])
}

func clear() {
	fmt.Fprint(screen, "\033[2J")
}

func draw(str string) {
	fmt.Fprint(screen, str)
}

func render() {
	screen.Flush()
}
