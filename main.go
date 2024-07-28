package main

import (
	"log"
	"math"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

var (
	screenWidth, screenHeight int
	grid                      [][]int
	styleGrid                 [][]int
	lastMouseX, lastMouseY    int
	mouseMoved                bool
)

func HSVtoRGB(hue int) (int32, int32, int32) {
	saturation := 1.0
	value := 1.0

	c := value * saturation
	x := c * (1 - math.Abs(math.Mod(float64(hue)/60.0, 2)-1))
	m := value - c

	var r, g, b float64

	switch {
	case hue >= 0 && hue < 60:
		r, g, b = c, x, 0
	case hue >= 60 && hue < 120:
		r, g, b = x, c, 0
	case hue >= 120 && hue < 180:
		r, g, b = 0, c, x
	case hue >= 180 && hue < 240:
		r, g, b = 0, x, c
	case hue >= 240 && hue < 300:
		r, g, b = x, 0, c
	case hue >= 300 && hue < 360:
		r, g, b = c, 0, x
	}

	r = (r + m) * 255
	g = (g + m) * 255
	b = (b + m) * 255

	return int32(r), int32(g), int32(b)
}

func render(s tcell.Screen) {

	for y := screenHeight - 2; y >= 0; y-- {
		for x := 0; x < screenWidth; x++ {
			blockstyle := tcell.StyleDefault.Background(tcell.NewRGBColor(HSVtoRGB(grid[y][x])))
			if grid[y][x] > 0 {
				if grid[y+1][x] == 0 {
					grid[y+1][x] = grid[y][x]
					grid[y][x] = 0
					s.SetContent(x, y+1, ' ', nil, blockstyle)
					styleGrid[y+1][x] = 1
				} else if x > 0 && grid[y+1][x-1] == 0 {
					grid[y+1][x-1] = grid[y][x]
					grid[y][x] = 0
					s.SetContent(x-1, y+1, ' ', nil, blockstyle)
					styleGrid[y+1][x-1] = 1
				} else if x < screenWidth-1 && grid[y+1][x+1] == 0 {
					grid[y+1][x+1] = grid[y][x]
					grid[y][x] = 0
					s.SetContent(x+1, y+1, ' ', nil, blockstyle)
					styleGrid[y+1][x+1] = 1
				}
			} else {
				style := tcell.StyleDefault
				if y != 0 && grid[y-1][x] != 0 {
					style = tcell.StyleDefault.Background(tcell.NewRGBColor(HSVtoRGB(grid[y-1][x])))
					grid[y][x] = grid[y-1][x]
					grid[y-1][x] = 0
					styleGrid[y][x] = 1
				} else {
					styleGrid[y][x] = 0
				}
				s.SetContent(x, y, ' ', nil, style)
			}
			if grid[y][x] != 0 && styleGrid[y][x] == 0 {
				s.SetContent(x, y, ' ', nil, tcell.StyleDefault.Background(tcell.NewRGBColor(HSVtoRGB(grid[y][x]))))
			}
		}
	}
}

func main() {
	s, err := tcell.NewScreen()
	if err != nil {
		log.Fatalf("TermISand couldn't initialise the screen :(")
	}
	defer s.Fini()

	if err := s.Init(); err != nil {
		log.Fatalf("TermISand couldn't initialise the screen :(")
	}

	s.EnableFocus()
	s.EnableMouse()
	screenWidth, screenHeight = s.Size()
	grid = make([][]int, screenHeight)
	styleGrid = make([][]int, screenHeight)

	for i := 0; i < screenHeight; i++ {
		grid[i] = make([]int, screenWidth)
		styleGrid[i] = make([]int, screenWidth)
	}

	s.Clear()
	colorNum := 0

	eventCh := make(chan tcell.Event, 1)

	go func() {
		for {
			ev := s.PollEvent()
			if ev != nil {
				eventCh <- ev
			}
		}
	}()

	ticker := time.NewTicker(20 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case ev := <-eventCh:
			switch ev := ev.(type) {
			case *tcell.EventKey:
				switch ev.Key() {
				case tcell.KeyCtrlQ:
					s.Fini()
					os.Exit(0)
				}
			case *tcell.EventResize:
				screenWidth, screenHeight = s.Size()
				s.Sync()
			case *tcell.EventMouse:
				lastMouseX, lastMouseY = ev.Position()
				mouseMoved = true
			}
		case <-ticker.C:
			// Add sand at the last known mouse position if the mouse has moved
			if mouseMoved && lastMouseY < screenHeight && lastMouseX < screenWidth && grid[lastMouseY][lastMouseX] == 0 {
				grid[lastMouseY][lastMouseX] = colorNum
				rand1 := rand.Intn(4)
				rand2 := rand.Intn(4)

				if (rand1 == 0 || rand2 == 0) && lastMouseY-1 >= 0 && grid[lastMouseY-1][lastMouseX] == 0 {
					grid[lastMouseY-1][lastMouseX] = colorNum
				}
				if (rand1 == 1 || rand2 == 1) && lastMouseY+1 < screenHeight && grid[lastMouseY+1][lastMouseX] == 0 {
					grid[lastMouseY+1][lastMouseX] = colorNum
				}
				if (rand1 == 2 || rand2 == 2) && lastMouseX-1 >= 0 && grid[lastMouseY][lastMouseX-1] == 0 {
					grid[lastMouseY][lastMouseX-1] = colorNum
				}
				if (rand1 == 3 || rand2 == 3) && lastMouseX+1 < screenWidth && grid[lastMouseY][lastMouseX+1] == 0 {
					grid[lastMouseY][lastMouseX+1] = colorNum
				}
			}

			render(s)
			s.Show()
			colorNum++
			if colorNum == 360 {
				colorNum = 1
			}
		}
	}
}
