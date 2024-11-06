package engine

import (
	"crypto/rand"
	"encoding/json"
	"fmt"
	"log/slog"
	"math/big"
	"os"
	"snake/internal/pkg/input"
	"snake/internal/pkg/player"

	"github.com/veandco/go-sdl2/img"
	"github.com/veandco/go-sdl2/sdl"
	"github.com/veandco/go-sdl2/ttf"
)

const (
	blank = iota
	snakeHead
	snakeBody
	snakeTail
	strawberry
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
	textures  map[string]*sdl.Texture
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
		textures:  make(map[string]*sdl.Texture),
	}

	for i := range e.grid {
		e.grid[i] = make([]int, gridWidth)
	}

	pRect.W = int32(e.player.Size)
	pRect.H = int32(e.player.Size)

	e.dropFruit(gridWidth, gridHeight)

	if err := e.LoadSpriteSheet(renderer, "./assets/snake.json", "./assets/snake.png"); err != nil {
		return Engine{}, fmt.Errorf("failed to create texture: %w", err)
	}

	return e, nil
}

type layer struct {
	Name string `json:"name"`
}

type size struct {
	W int `json:"w"`
	H int `json:"h"`
}

type meta struct {
	Image     string  `json:"image"`
	Format    string  `json:"format"`
	SheetSize size    `json:"size"`
	Scale     string  `json:"scale"`
	Layers    []layer `json:"layers"`
}

type spritesheet struct {
	Meta meta `json:"meta"`
}

func (e *Engine) LoadSpriteSheet(renderer *sdl.Renderer, jsonFile, assetFile string) error {
	data, err := os.ReadFile(jsonFile)
	if err != nil {
		return fmt.Errorf("error reading data file", err)
	}

	var spriteSheetData spritesheet

	err = json.Unmarshal(data, &spriteSheetData)
	if err != nil {
		return fmt.Errorf("error Unmarshalling data", err)
	}

	surface, err := img.Load(assetFile)
	if err != nil {
		return fmt.Errorf("error adding texture to asset store: %w", err)
	}

	spriteTexture, err := renderer.CreateTextureFromSurface(surface)
	if err != nil {
		return fmt.Errorf("error adding texture to asset store: %w", err)
	}
	surface.Free()

	srcRect := sdl.Rect{
		W: 25,
		H: 25,
	}
	for i, frame := range spriteSheetData.Meta.Layers {
		t, err := renderer.CreateTexture(sdl.PIXELFORMAT_RGBA8888, sdl.TEXTUREACCESS_TARGET, 25, 25)
		if err != nil {
			return fmt.Errorf("error creating frame texture: %w", err)
		}

		srcRect.X = int32(i * int(srcRect.W))
		err = renderer.SetRenderTarget(t)
		if err != nil {
			return fmt.Errorf("error setting render target: %w", err)
		}

		err = renderer.Copy(spriteTexture, &srcRect, nil)
		if err != nil {
			return fmt.Errorf("error copying texture: %w", err)
		}

		e.textures[frame.Name] = t
	}

	err = renderer.SetRenderTarget(nil)
	if err != nil {
		return fmt.Errorf("error setting render target: %w", err)
	}

	spriteTexture.Destroy()

	return nil
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
	// TODO: figure out transparency
	err := e.renderer.SetDrawColor(0, 0, 0, 255)
	if err != nil {
		slog.Error("engine render", "error", err)
	}
	err = e.renderer.Clear()
	if err != nil {
		slog.Error("engine render clear", "error", err)
	}

	for y, r := range e.grid {
		for x, v := range r {
			pRect.X = int32(x * e.player.Size)
			pRect.Y = int32(y * e.player.Size)

			var tex *sdl.Texture
			var found bool
			rotation := float64(0)

			if v != blank {
				switch v {
				case snakeHead:
					tex, found = e.textures["snake-head"]
					if !found {
						slog.Error("engine texture", "error", "snake texture not found")
					}
					if e.player.Direction[0] == input.Left {
						rotation = 180
					} else if e.player.Direction[0] == input.Up {
						rotation = 270
					} else if e.player.Direction[0] == input.Down {
						rotation = 90
					}

				case snakeBody:
					tex, found = e.textures["snake-body"]
					if !found {
						slog.Error("engine texture", "error", "snake body texture not found")
					}

				case snakeTail:
					tex, found = e.textures["snake-tail"]
					if !found {
						slog.Error("engine texture", "error", "snake tail texture not found")
					}
					if e.player.Body[len(e.player.Body)-1].X == e.player.Body[len(e.player.Body)-2].X {
						if e.player.Direction[0] == input.Down {
							rotation = 90
						} else {
							rotation = 270
						}
					} else {
						if e.player.Direction[0] == input.Left {
							rotation = 180
						}
					}

				case strawberry:
					tex, found = e.textures["strawberry"]
					if !found {
						slog.Error("engine texture", "error", "strawberry texture not found")
					}
				}

				// err := e.renderer.FillRect(&pRect)
				// if err != nil {
				// 	slog.Error("engine fillRect", "error", err)
				// }

				if err := e.renderer.CopyEx(tex, nil, &pRect, rotation, nil, 0); err != nil {
					slog.Error("engine texture render", "error", err)
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
	e.grid[y][x] = strawberry
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
