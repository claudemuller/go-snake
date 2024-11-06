package player

import (
	"snake/internal/pkg/input"
)

const moveIntervalSpeed = 1000.0

type Pos struct{ X, Y int }

type Player struct {
	Direction    [2]int
	Size         int
	length       int
	Body         []Pos
	moveInterval float32
	speed        float32
}

func New(winW, winH int) *Player {
	p := Player{
		Direction:    [2]int{input.Right, input.Right},
		Size:         25,
		length:       5,
		moveInterval: 500.0,
		speed:        300.0,
	}
	// gridW := winW / p.Size
	// gridH := winH / p.Size

	p.Body = make([]Pos, p.length)
	for i := range p.Body {
		p.Body[i] = Pos{X: 0, Y: 0} // gridW/2 + i, Y: gridH / 2}
	}

	return &p
}

const (
	blank = iota
	snakeHead
	snakeBody
	snakeTail
	strawberry
)

func (p *Player) Update(delta float32, grid [][]int, fruit []Pos, collisionFn func(int, int)) {
	p.moveInterval += delta * moveIntervalSpeed
	gridHeight := len(grid)
	gridWidth := len(grid[0])

	if p.moveInterval >= p.speed {
		head := &p.Body[0]
		tail := p.Body[len(p.Body)-1]

		grid[tail.Y][tail.X] = blank

		// Move segments up
		for i := len(p.Body) - 1; i > 0; i-- {
			p.Body[i] = p.Body[i-1]
		}

		switch p.Direction[0] {
		case input.Up:
			// if p.Body[1].X == head.X {
			// 	p.Direction[0] = p.Direction[1]
			// 	break
			// }
			head.Y = (head.Y - 1) % gridHeight
			if head.Y < 0 {
				head.Y += gridHeight
			}

		case input.Down:
			// if p.Body[1].X == head.X {
			// 	p.Direction[0] = p.Direction[1]
			// 	break
			// }
			head.Y = (head.Y + 1) % gridHeight

		case input.Left:
			// if p.Body[1].Y == head.Y {
			// 	p.Direction[0] = p.Direction[1]
			// 	break
			// }
			head.X = (head.X - 1) % gridWidth
			if head.X < 0 {
				head.X += gridWidth
			}

		case input.Right:
			// if p.Body[1].Y == head.Y {
			// 	p.Direction[0] = p.Direction[1]
			// 	break
			// }
			head.X = (head.X + 1) % gridWidth
		}

		// Check collisions between player and fruit
		for _, f := range fruit {
			if f.X == head.X && f.Y == head.Y {
				collisionFn(gridWidth, gridHeight)
				p.length += 1
				p.Body = append(p.Body, Pos{X: tail.X, Y: tail.Y})
			}
		}

		// Place snake into grid
		for i, segment := range p.Body {
			switch i {
			case 0:
				grid[segment.Y][segment.X] = snakeHead

			case len(p.Body) - 1:
				grid[segment.Y][segment.X] = snakeTail

			default:
				grid[segment.Y][segment.X] = snakeBody
			}
		}

		p.moveInterval = 0
	}
}
