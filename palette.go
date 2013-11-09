package main

// Poorly named function to convert a (r,g,b) tuple of floats [0.0...1.0] to a 32bit packed RGB value
func floatToInt(r,g,b float64) uint32 {
	sr, sg, sb := uint32(r*255.0), uint32(g*255.0), uint32(b*255.0)
	if sr > 255 {
		sr = 255
	}
	if sg > 255 {
		sg = 255
	}
	if sb > 255 {
		sb = 255
	}
	return uint32(sr << 16 | sg << 8 | sb)
}

// Iterate between cur and dest rgb values and output into the given palette
func doColorSteps(cur_r, cur_g, cur_b, dest_r, dest_g, dest_b float64, pal[] uint32, start,end int) {
	steps := float64(end - start + 1)
	delta_r := (dest_r - cur_r) / steps
	delta_g := (dest_g - cur_g) / steps
	delta_b := (dest_b - cur_b) / steps
	
	for i := start; i<= end; i++ {
		pal[i] = floatToInt(cur_r, cur_g, cur_b)
		cur_r += delta_r
		cur_g += delta_g
		cur_b += delta_b
	}
}

// create a hardwired palette
func NewPalette () []uint32 {
    const MAX = 400
    colors := make([]uint32, MAX)

    doColorSteps(0.0, 0.0, 0.7, 1.0, 0.0, 0.0, colors, 0, 99)
    doColorSteps(1.0, 0.0, 0.0, 0.0, 1.0, 0.0, colors, 100, 199)
    doColorSteps(0.0, 1.0, 0.0, 1.0, 1.0, 0.0, colors, 200, 299)
    doColorSteps(1.0, 1.0, 0.0, 0.0, 0.0, 0.7, colors, 300, 399)
    
    return colors
}
