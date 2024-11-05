package player

import (
	"snake/internal/pkg/colours"
	"snake/internal/pkg/input"
)

const moveIntervalSpeed = 1000.0

type Pos struct{ X, Y int }

type Player struct {
	Direction    [2]int
	Size         int
	length       int
	body         []Pos
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
	gridW := winW / p.Size
	gridH := winH / p.Size

	p.body = make([]Pos, p.length)
	for i := range p.body {
		p.body[i] = Pos{X: gridW/2 + i, Y: gridH / 2}
	}

	return &p
}

func (p *Player) Update(delta float32, grid [][]int, fruit []Pos, collisionFn func(int, int)) {
	p.moveInterval += delta * moveIntervalSpeed
	gridHeight := len(grid)
	gridWidth := len(grid[0])

	if p.moveInterval >= p.speed {
		head := &p.body[0]
		tail := p.body[len(p.body)-1]

		grid[tail.Y][tail.X] = colours.Blank

		// Move segments up
		for i := len(p.body) - 1; i > 0; i-- {
			p.body[i] = p.body[i-1]
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
				p.body = append(p.body, Pos{X: tail.X, Y: tail.Y})
			}
		}

		// Place snake into grid
		for _, segment := range p.body {
			grid[segment.Y][segment.X] = colours.White
		}

		p.moveInterval = 0
	}
}
