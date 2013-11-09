package main

import (
	"fmt"
	"github.com/jonhanks/Go-SDL/sdl"
	"log"
	"runtime"
	"time"
)

// Basic size information
type Size struct {
	Width, Height int
}

// run the fractal rendering
func doFractal(state *FractalData, screen *sdl.Surface, pal []uint32, done chan bool) {
	f := func() {
		done <- true
	}
	defer f()
	start := time.Now()
	Mandelbrot(state, screen, pal)
	delta := time.Since(start)
	fmt.Println("Compute and Render done in ", delta)
}

func main() {
	WIDTH := 640
	HEIGHT := 480
	size := Size{Width: WIDTH, Height: HEIGHT}

	runtime.GOMAXPROCS(runtime.NumCPU() * 2)

	if sdl.Init(sdl.INIT_EVERYTHING) != 0 {
		log.Fatal(sdl.GetError())
	}
	defer sdl.Quit()

	var screen = sdl.SetVideoMode(WIDTH, HEIGHT, 32, sdl.RESIZABLE)

	if screen == nil {
		log.Fatal(sdl.GetError())
	}

	var video_info = sdl.GetVideoInfo()

	println("HW_available = ", video_info.HW_available)
	println("WM_available = ", video_info.WM_available)
	println("Video_mem = ", video_info.Video_mem, "kb")

	sdl.EnableUNICODE(1)

	sdl.WM_SetCaption("Go-SDL SDL Test", "")

	running := true
	redraw := true
	requestRedraw := false
	redrawFlag := &redraw

	fullState := NewFractalData(WIDTH, HEIGHT, 2.0, 200)
	fullState.Center = complex(-0.5, 0.0)

	state := fullState.FractalState
	state.DebugRender = true
	state.DebugMerge = true
	pal := NewPalette()

	keymap := make(map[string]bool)

	if sdl.GetKeyName(270) != "[+]" {
		log.Fatal("GetKeyName broken")
	}

	// Note: The following SDL code is highly ineffective.
	//       It is eating too much CPU. If you intend to use Go-SDL,
	//       you should to do better than this.

	fractalReady := make(chan bool)

	for running {
		if redraw {
			fmt.Println("Parsing keypresses")
			if keymap["+"] {
				state.Scale *= 0.8
			}
			if keymap["-"] {
				state.Scale *= 1.2
			}
			if keymap["left"] {
				state.Center -= complex(state.Scale*0.1, 0.0)
			}
			if keymap["right"] {
				state.Center += complex(state.Scale*0.1, 0.0)
			}
			if keymap["up"] {
				state.Center += complex(0.0, state.Scale*0.1)
			}
			if keymap["down"] {
				state.Center -= complex(0.0, state.Scale*0.1)
			}
			if keymap["iterup"] {
				state.MaxIterations += 25
				if state.MaxIterations > 2000 {
					state.MaxIterations = 2000
				}
			}
			if keymap["iterdown"] {
				state.MaxIterations -= 25
				if state.MaxIterations < 100 {
					state.MaxIterations = 100
				}
			}
			if keymap["overlay"] {
				state.DebugOverlay = !state.DebugOverlay
			}
			keymap = make(map[string]bool)
			fmt.Println("Queue fractal")
			screen.FillRect(nil, 0)
			if size.Width == WIDTH && size.Height == HEIGHT {
				fullState.Merge(state)
			} else {
				// resize event
				screen = sdl.SetVideoMode(size.Width, size.Height, 32, sdl.RESIZABLE)

				if screen == nil {
					log.Fatal(sdl.GetError())
				}
				WIDTH = size.Width
				HEIGHT = size.Height
				state.Width = WIDTH
				state.Height = HEIGHT
				fullState = NewFractalData(WIDTH, HEIGHT, state.Scale, state.MaxIterations)
				fullState.FractalState = state
			}
			fmt.Println("Iterations: ", fullState.MaxIterations)
			go doFractal(fullState, screen, pal, fractalReady)
			redraw = false
			requestRedraw = false
			redrawFlag = &requestRedraw
		}
		select {

		case _ = <-fractalReady:
			fmt.Println("Fractal reports ready")
			screen.Flip()
			redraw = requestRedraw
			requestRedraw = false
			redrawFlag = &redraw
		case _event := <-sdl.Events:
			switch e := _event.(type) {
			case sdl.QuitEvent:
				running = false

			case sdl.KeyboardEvent:
				println("")
				println(e.Keysym.Sym, ": ", sdl.GetKeyName(sdl.Key(e.Keysym.Sym)))

				if e.Keysym.Sym == sdl.K_ESCAPE || e.Keysym.Sym == sdl.K_q {
					running = false
				}
				if e.Type == sdl.KEYDOWN {
					if e.Keysym.Sym == sdl.K_KP_PLUS || e.Keysym.Sym == sdl.K_EQUALS {
						fmt.Println("zoom in")
						keymap["+"] = true
						*redrawFlag = true
					} else if e.Keysym.Sym == sdl.K_KP_MINUS || e.Keysym.Sym == sdl.K_MINUS {
						fmt.Println("zoom out")
						keymap["-"] = true
						*redrawFlag = true
					} else if e.Keysym.Sym == sdl.K_LEFT {
						keymap["left"] = true
						*redrawFlag = true
					} else if e.Keysym.Sym == sdl.K_RIGHT {
						keymap["right"] = true
						*redrawFlag = true
					} else if e.Keysym.Sym == sdl.K_UP {
						keymap["up"] = true
						*redrawFlag = true
					} else if e.Keysym.Sym == sdl.K_DOWN {
						keymap["down"] = true
						*redrawFlag = true
					} else if e.Keysym.Sym == sdl.K_LEFTBRACKET {
						keymap["iterup"] = true
						*redrawFlag = true
					} else if e.Keysym.Sym == sdl.K_RIGHTBRACKET {
						keymap["iterdown"] = true
						*redrawFlag = true
					} else if e.Keysym.Sym == sdl.K_o {
						keymap["overlay"] = true
						*redrawFlag = true
					}
				}
				/*fmt.Printf("%04x ", e.Type)

				for i := 0; i < len(e.Pad0); i++ {
					fmt.Printf("%02x ", e.Pad0[i])
				}
				println()

				fmt.Printf("Type: %02x Which: %02x fullState: %02x Pad: %02x\n", e.Type, e.Which, e.State, e.Pad0[0])
				fmt.Printf("Scancode: %02x Sym: %08x Mod: %04x Unicode: %04x\n", e.Keysym.Scancode, e.Keysym.Sym, e.Keysym.Mod, e.Keysym.Unicode)
				*/
			case sdl.MouseButtonEvent:
				if e.Type == sdl.MOUSEBUTTONDOWN {
					println("Click:", e.X, e.Y)
				}

			case sdl.ResizeEvent:
				println("resize screen ", e.W, e.H)
				//panic("Resize not supported yet")

				size.Width = int(e.W)
				size.Height = int(e.H)
				*redrawFlag = true

				//screen = sdl.SetVideoMode(int(e.W), int(e.H), 32, sdl.RESIZABLE)

				//if screen == nil {
				//	log.Fatal(sdl.GetError())
				//}
			}
		}
	}
}
