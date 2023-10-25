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

type food struct {
	color    string
	shape    string
	level    int
	position position
}

type game struct {
	score int
	snake *snake
	foods []food
	// food  food
}

type snake struct {
	body          []food
	prevDirection direction
	direction     direction
}

type position [2]int

type direction int

const (
	north direction = iota
	east
	south
	west
)

var foods []food

func init() {
	// 创建或打开日志文件
	logFile, err := os.OpenFile("run.log", os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal("无法打开日志文件:", err)
	}
	// defer logFile.Close()
	// 设置日志输出到文件
	log.SetOutput(logFile)

	rand.New(rand.NewSource(time.Now().UnixNano()))
	foods = append(foods, food{
		color: cYellow,
		shape: "*",
		level: 1,
	}, food{
		color: cBlue,
		shape: "#",
		level: 2,
	}, food{
		color: cGreen,
		shape: "%",
		level: 3,
	})
}

func (g *game) randomFood() food {
	randomIndex := rand.Intn(len(foods))
	pos := randomPosition()
	food := foods[randomIndex]
	food.position = pos
	return food
}

func newGame() *game {
	log.Println("game init")
	snake := newSnake()

	game := &game{
		score: 0,
		snake: snake,
	}
	game.foods = append(game.foods, game.randomFood())
	go game.listenForKeyPress()
	return game
}

func newSnake() *snake {
	log.Println("snake init")
	maxX, maxY := getSize()
	pos := position{maxX / 2, maxY / 2}
	log.Printf("snake pos:%d,%d", pos[0], pos[1])
	snake := &snake{
		direction: north,
	}
	snake.body = append(snake.body, food{
		color:    cWhite,
		shape:    "O",
		level:    0,
		position: pos,
	},
	)
	return snake

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
			if game.snake.prevDirection == south {
				newHeadPos.position[0]++
				game.snake.prevDirection = game.snake.direction
			}
			newHeadPos.position[1]--
		case east:
			newHeadPos.position[0]++
		case south:
			if game.snake.prevDirection == north {
				newHeadPos.position[0]++
				game.snake.prevDirection = game.snake.direction
			}
			newHeadPos.position[1]++
		case west:
			newHeadPos.position[0]--

		}
		log.Printf("snake prev direction:%d", game.snake.prevDirection)
		log.Printf("snake current direction:%d", game.snake.direction)

		hitWall := newHeadPos.position[0] < 1 || newHeadPos.position[1] < 1 || newHeadPos.position[0] > x ||
			newHeadPos.position[1] > y
		if hitWall {
			log.Println("hit wall")
			game.over()
		}
		for _, pos := range game.snake.body {
			if positionsAreSame(newHeadPos.position, pos.position) {
				log.Printf("current pos:%d,%d", newHeadPos.position[0], newHeadPos.position[1])
				log.Printf("tick pos pos:%d,%d", pos.position[0], pos.position[1])
				log.Println("eat yourself")
				game.over()
			}
		}

		game.snake.body = append([]food{newHeadPos}, game.snake.body...)
		log.Println(game.snake.body)

		ateFood := game.matchFood(newHeadPos.position)
		if ateFood.level > 0 {
			log.Println("ateFood")
			log.Println(ateFood)
			ateFood.position = newHeadPos.position
			// game.snake.body = append([]food{ateFood}, game.snake.body...)
			game.score = game.score + ateFood.level
			game.placeNewFood()
			game.snake.body[0].color = ateFood.color
			game.snake.body[0].shape = ateFood.shape
		} else {
			log.Println("less snake")
			log.Println(game.snake.body)
			game.snake.body = game.snake.body[:len(game.snake.body)-1]
		}
		game.draw()

	}
}

func (g *game) placeNewFood() {
	for {
		if len(g.foods) > 2 {
			continue
		}
		newFood := g.randomFood()
		// newFoodPosition := randomPosition()

		for _, v := range g.foods {
			if positionsAreSame(newFood.position, v.position) {
				continue
			}
		}

		// if positionsAreSame(newFoodPosition, g.food) {
		// 	continue
		// }

		for _, pos := range g.snake.body {
			if positionsAreSame(newFood.position, pos.position) {
				continue
			}
		}
		g.foods = append(g.foods, newFood)
		break
	}
}

func (g *game) matchFood(b position) (f food) {
	for i := len(g.foods) - 1; i >= 0; i-- {
		if g.foods[i].position[0] == b[0] && g.foods[i].position[1] == b[1] {
			f = g.foods[i]
			g.foods = append(g.foods[:i], g.foods[i+1:]...)
			return
		}
	}
	return food{}
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
	for _, v := range g.foods {
		moveCursor(v.position)
		draw(v.color + v.shape)
	}

	log.Printf("draw: %v", g.snake.body)
	for _, pos := range g.snake.body {
		moveCursor(pos.position)
		draw(pos.color + pos.shape)
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
	x := rand.Intn(width-10) + 1
	y := rand.Intn(height-10) + 2

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
			g.snake.prevDirection = g.snake.direction
			g.snake.direction = north
		case 'B':
			g.snake.prevDirection = g.snake.direction
			g.snake.direction = south
		case 'C':
			g.snake.prevDirection = g.snake.direction
			g.snake.direction = east
		case 'D':
			g.snake.prevDirection = g.snake.direction
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
