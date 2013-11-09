package main

import (
	"fmt"
	"github.com/jonhanks/Go-SDL/sdl"
	"math/cmplx"
	"time"
	"unsafe"
	"sync"
)

// A sample point in the fractal space
type FractalSample struct {
	Value     complex128		// currently computed value
	Escape    int 				// number of iterations prior to escape
	Reuse     bool 				// should this sample be reused
	Recompute bool				// should this sample be recomputed
}

// Global state of a fractal rendering
type FractalState struct {
	Width, Height int 			// Dimensions to render the image as
	MaxIterations int 			// Number of iterations
	Scale         float64 		// Zoom
	Center        complex128	// Center point of the view
	DebugOverlay  bool			// debugging flags
	DebugCompute  bool
	DebugRender   bool
	DebugMerge    bool
}

// Return the aspect ratio of the output
func (state *FractalState) Aspect() float64 {
	return float64(state.Width) / float64(state.Height)
}

// Return the values that x/y are incremented by when moving 1 row or 1 column
func (state *FractalState) Increments() (x_incr, y_incr float64) {
	x_incr = state.Scale / float64(state.Width)
	y_incr = (state.Scale / state.Aspect()) / float64(state.Height)
	return
}

// The fractal state + samples
type FractalData struct {
	FractalState
	Samples [][]FractalSample
}

// Attempt to merge a new FractalState with the current FractalData (state + samples).
// This sets the reuse/recompute bits on the Samples[][].
// This function is provided to allow optimizations.
func (fd *FractalData) Merge(state FractalState) {
	start := time.Now()
	if fd.Scale == state.Scale && fd.Width == state.Width && fd.Height == state.Height {
		if fd.Center == state.Center {
			// Check for a change in iterations
			if fd.MaxIterations != state.MaxIterations {
				fmt.Println("REUSING DATA OPTIMIZING FOR ITERATION CHANGE")
				reuseCnt, recompCnt := 0, 0
				// iteration value change
				if fd.MaxIterations < state.MaxIterations {
					// iterations go up, reuse everything
					for y, row := range fd.Samples {
						for x, _ := range row {
							if fd.Samples[y][x].Escape < fd.MaxIterations {
								fd.Samples[y][x].Reuse = true
								reuseCnt++
							} else {
								fd.Samples[y][x].Recompute = true
								recompCnt++
							}
						}
					}
				} else {
					// iterations go down, reuse everything with Escape < state.MaxIterations
					for y, row := range fd.Samples {
						for x, _ := range row {
							if fd.Samples[y][x].Escape > state.MaxIterations {
								fd.Samples[y][x].Escape = state.MaxIterations
							}
							fd.Samples[y][x].Reuse = true
							reuseCnt++
						}
					}
				}
				fmt.Println("Scheduled to reuse ", reuseCnt, " and recompute ", recompCnt)
			}
		} else {
			delta := state.Center - fd.Center
			x_incr, y_incr := fd.FractalState.Increments()
			delta_x := int(real(delta) / x_incr)
			delta_y := -int(imag(delta) / y_incr)

			fmt.Println("Translated by ", delta_x, " and ", delta_y)

			// currently delta_x = - for left, + for right
			// delta_y = - for up, + for down
			// directions that where requested, opposite of how the fractal moves

			if abs(delta_x) < state.Width/2 && abs(delta_y) < state.Height/2 {
				// fixme this fails with there is change in x & y
				// Y first
				if delta_y != 0 {
					if delta_y < 0 {
						delta_y = -delta_y
						// move image down dela_y rows
						for y := fd.Height - delta_y - 1; y >= 0; y-- {
							target_y := y + delta_y
							for x := 0; x < fd.Width; x++ {
								fd.Samples[target_y][x] = fd.Samples[y][x]
								fd.Samples[target_y][x].Reuse = true
							}
						}
					} else {
						// move image up dela_y rows
						for y := delta_y; y < fd.Height; y++ {
							target_y := y - delta_y
							for x := 0; x < state.Width; x++ {
								fd.Samples[target_y][x] = fd.Samples[y][x]
								fd.Samples[target_y][x].Reuse = true
							}
						}
					}
				}

				// then X
				if delta_x != 0 {
					if delta_x < 0 {
						delta_x = -delta_x
						// move the image right delta_x rows
						for y := 0; y < fd.Height; y++ {
							for x := fd.Width - delta_x - 1; x >= 0; x-- {
								target_x := x + delta_x
								fd.Samples[y][target_x] = fd.Samples[y][x]
								fd.Samples[y][target_x].Reuse = true
							}
						}
					} else {
						// move the image left detla_x rows
						for y := 0; y < fd.Height; y++ {
							for x := delta_x; x < fd.Width; x++ {
								target_x := x - delta_x
								fd.Samples[y][target_x] = fd.Samples[y][x]
								fd.Samples[y][target_x].Reuse = true
							}
						}
					}
				}
			}
		}

	}
	fd.FractalState = state
	if state.DebugMerge {
		fmt.Println("Merge took ", time.Since(start))
	}
}

// Create a FractalData object.
func NewFractalData(width, height int, scale float64, maxIterations int) *FractalData {
	state := &FractalData{FractalState: FractalState{Width: width, Height: height, Scale: scale, MaxIterations: maxIterations}, Samples: make([][]FractalSample, height)}
	for i := range state.Samples {
		state.Samples[i] = make([]FractalSample, width)
	}
	return state
}

type pixelPtr *uint32

// compute the mandelbrot function
func mandelbrot_f(c, z0 complex128, curIterations, maxIterations int) FractalSample {
	z := z0

	i := curIterations
	for ; i < maxIterations && cmplx.Abs(z) < 2.0; i++ {
		z = z*z + c
	}
	return FractalSample{Value: z, Escape: i}
}

// compute 1 row of mandlebrot functions, taking into account the reuse and recompute bits on the samples
func mandelbrot_row(x_cur float64, y_cur float64, xincr float64, y int, state *FractalData, wg *sync.WaitGroup) {
	defer wg.Done()
	if state.DebugCompute {
		fmt.Print("#")
	}
	for x := 0; x < state.Width; x++ {
		if !state.Samples[y][x].Reuse {
			c := complex(x_cur, y_cur)
			if !state.Samples[y][x].Recompute {
				state.Samples[y][x] = mandelbrot_f(c, c, 0, state.MaxIterations)
			} else {
				state.Samples[y][x] = mandelbrot_f(c, state.Samples[y][x].Value, state.Samples[y][x].Escape, state.MaxIterations)
				state.Samples[y][x].Recompute = true
			}
		} else {
			state.Samples[y][x].Reuse = true
		}
		x_cur += xincr
	}
}

// Render a row into a 32bit bitmap.
// curOffset - offset into the bitmap
// y - row value
// state - the Fractal information and samples
// pal - an slice of 32bit color values which will be indexed by escape values
// c - channel used to signal completion
func renderRow(curOffset uintptr, y int, state *FractalData, pal []uint32, wg *sync.WaitGroup) {
	defer wg.Done()
	lp := len(pal)
	if state.DebugOverlay {
		for x := 0; x < state.Width; x++ {
			if state.Samples[y][x].Reuse {
				*pixelPtr(unsafe.Pointer(curOffset)) = uint32(0x00ff0000)
			} else if state.Samples[y][x].Recompute {
				*pixelPtr(unsafe.Pointer(curOffset)) = uint32(0x0000ff00)
			} else {
				escape := state.Samples[y][x].Escape
				if escape < state.MaxIterations {
					*pixelPtr(unsafe.Pointer(curOffset)) = pal[escape%lp]
				}
			}
			state.Samples[y][x].Reuse, state.Samples[y][x].Recompute = false, false
			curOffset += 4
		}
	} else {
		for x := 0; x < state.Width; x++ {
			escape := state.Samples[y][x].Escape
			if escape < state.MaxIterations {
				*pixelPtr(unsafe.Pointer(curOffset)) = pal[escape%lp]
			}
			state.Samples[y][x].Reuse, state.Samples[y][x].Recompute = false, false
			curOffset += 4
		}
	}
}

// Render a mandelbrot fractal onto the given sdl surface
func Mandelbrot(state *FractalData, screen *sdl.Surface, pal []uint32) {
	fmt.Println("Beginning Mandelbrot")
	var wg sync.WaitGroup
	aspect := state.FractalState.Aspect()
	xincr, yincr := state.FractalState.Increments()
	y_cur := imag(state.Center) + (state.Scale/aspect)/2.0

	wg.Add(state.Height)
	for y := 0; y < state.Height; y++ {
		x_cur := real(state.Center) - state.Scale/2.0
		go mandelbrot_row(x_cur, y_cur, xincr, y, state, &wg)
		y_cur -= yincr
	}
	wg.Wait()
	fmt.Println("Rendering mandelbrot")

	screen.Lock()
	defer screen.Unlock()

	wg.Add(state.Height)
	offset := uintptr(unsafe.Pointer(screen.Pixels))
	delta := uintptr(screen.Pitch / 1)
	for y := 0; y < state.Height; y++ {
		go renderRow(offset, y, state, pal, &wg)
		offset += delta
	}
	wg.Wait()
	//invocations++
	fmt.Println("Mandeblrot completed")
}

// return |val|
func abs(val int) int {
	if val < 0 {
		return -val
	}
	return val
}
