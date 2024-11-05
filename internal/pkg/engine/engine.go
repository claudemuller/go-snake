package engine

import (
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"snake/internal/pkg/colours"
	"snake/internal/pkg/input"
	"snake/internal/pkg/player"

	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	fps               = 30
	millisecsPerFrame = 1000.0 / fps
)

var prevFrameMS float32

type Engine struct {
	window    *sdl.Window
	renderer  *sdl.Renderer
	grid      [][]int
	player    *player.Player
	fruit     []player.Pos
	isRunning bool
	width     int32
	height    int32
}

func New(title string, winWidth, winHeight int32) (Engine, error) {
	if err := sdl.Init(sdl.INIT_EVERYTHING); err != nil {
		return Engine{}, fmt.Errorf("error initialising SDL: %v", err)
	}

	if err := ttf.Init(); err != nil {
		return Engine{}, fmt.Errorf("error initialising TTF: %v", err)
	}

	window, err := sdl.CreateWindow(title, sdl.WINDOWPOS_UNDEFINED, sdl.WINDOWPOS_UNDEFINED, winWidth, winHeight, sdl.WINDOW_SHOWN)
	if err != nil {
		return Engine{}, fmt.Errorf("error creating window: %v", err)
	}

	renderer, err := sdl.CreateRenderer(window, -1, sdl.RENDERER_ACCELERATED|sdl.RENDERER_PRESENTVSYNC)
	if err != nil {
		return Engine{}, fmt.Errorf("error creating renderer: %v", err)
	}

	gridWidth := 40
	gridHeight := 30
	e := Engine{
		isRunning: true,
		width:     winWidth,
		height:    winHeight,
		window:    window,
		renderer:  renderer,
		player:    player.New(int(winWidth), int(winHeight)),
		grid:      make([][]int, gridHeight),
	}

	for i := range e.grid {
		e.grid[i] = make([]int, gridWidth)
	}

	pRect.W = int32(e.player.Size)
	pRect.H = int32(e.player.Size)

	e.dropFruit(gridWidth, gridHeight)

	return e, nil
}

func (e *Engine) Run() {
	for e.isRunning {
		e.processInput()

		currentTicks := float32(sdl.GetTicks64())
		timeToWait := millisecsPerFrame - (currentTicks - prevFrameMS)

		if timeToWait > 0 && timeToWait <= millisecsPerFrame {
			sdl.Delay(uint32(timeToWait))
		}

		delta := (currentTicks - prevFrameMS) / 1000.0
		prevFrameMS = currentTicks

		e.update(delta)
		e.render()
	}

}

func (e *Engine) processInput() {
	for event := sdl.PollEvent(); event != nil; event = sdl.PollEvent() {
		switch ev := event.(type) {
		case *sdl.QuitEvent:
			e.isRunning = false

		case *sdl.KeyboardEvent:
			if ev.Keysym.Sym == sdl.K_ESCAPE {
				e.isRunning = false
			}

			e.player.Direction[1] = e.player.Direction[0]
			if ev.Keysym.Sym == sdl.K_UP {
				e.player.Direction[0] = input.Up
			}
			if ev.Keysym.Sym == sdl.K_DOWN {
				e.player.Direction[0] = input.Down
			}
			if ev.Keysym.Sym == sdl.K_LEFT {
				e.player.Direction[0] = input.Left
			}
			if ev.Keysym.Sym == sdl.K_RIGHT {
				e.player.Direction[0] = input.Right
			}
		}
	}
}

func (e *Engine) update(delta float32) {
	e.player.Update(delta, e.grid, e.fruit, e.dropFruit)
}

var pRect sdl.Rect

func (e *Engine) render() {
	err := e.renderer.SetDrawColor(21, 21, 21, 255)
	if err != nil {
		slog.Error("engine render", "error", err)
	}
	err = e.renderer.Clear()
	if err != nil {
		slog.Error("engine render clear", "error", err)
	}

	for y, r := range e.grid {
		for x, v := range r {
			pRect.X = int32(x) * int32(e.player.Size)
			pRect.Y = int32(y) * int32(e.player.Size)

			if v != colours.Blank {
				switch v {
				case colours.White:
					err := e.renderer.SetDrawColor(255, 255, 255, 255)
					if err != nil {
						slog.Error("engine render", "error", err)
					}

				case colours.Yellow:
					err := e.renderer.SetDrawColor(255, 255, 0, 255)
					if err != nil {
						slog.Error("engine render", "error", err)
					}
				}

				err := e.renderer.FillRect(&pRect)
				if err != nil {
					slog.Error("engine fillRect", "error", err)
				}
			}

			// testGrid(e.renderer)
		}
	}

	e.renderer.Present()
}

func testGrid(renderer *sdl.Renderer) {
	r := uint8(genRandInt(255))
	g := uint8(genRandInt(255))
	b := uint8(genRandInt(255))
	renderer.SetDrawColor(r, g, b, 255)
	renderer.FillRect(&pRect)

}

func (e *Engine) dropFruit(gridWidth, gridHeight int) {
	x := genRandInt(gridWidth)
	y := genRandInt(gridHeight)
	e.grid[y][x] = colours.Yellow
	e.fruit = append(e.fruit, player.Pos{X: x, Y: y})
}

func (e *Engine) Cleanup() {
	err := e.renderer.Destroy()
	if err != nil {
		slog.Error("engine destroy", "error", err)
	}

	err = e.window.Destroy()
	if err != nil {
		slog.Error("engine destroy", "error", err)
	}

	// ttf.Quit()
	sdl.Quit()
}

func genRandInt(nMax int) int {
	max := big.NewInt(int64(nMax))
	n, err := rand.Int(rand.Reader, max)
	if err != nil {
		slog.Error("engine dropFruit", "error", err)
		return 0
	}
	return int(n.Int64())
}
