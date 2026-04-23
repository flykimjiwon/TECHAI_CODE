// +build ignore

// gen-icon generates TECHAI CODE app icons (1024x1024 PNG).
// Usage: go run scripts/gen-icon.go
package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"math"
	"os"
)

func main() {
	generateIcon("techai-tui-icon.png", true)
	generateIcon("techai-ide-icon.png", false)
	println("Generated: techai-tui-icon.png, techai-ide-icon.png")
}

func generateIcon(filename string, isTUI bool) {
	size := 1024
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	// Background - rounded square with gradient
	bgTop := color.RGBA{15, 23, 42, 255}    // slate-900
	bgBot := color.RGBA{30, 41, 59, 255}    // slate-800
	if !isTUI {
		bgTop = color.RGBA{10, 10, 12, 255}  // IDE darker
		bgBot = color.RGBA{17, 17, 21, 255}
	}

	// Fill background with vertical gradient
	for y := 0; y < size; y++ {
		t := float64(y) / float64(size)
		c := lerpColor(bgTop, bgBot, t)
		for x := 0; x < size; x++ {
			// Rounded corners (radius 180)
			if isInRoundedRect(x, y, size, size, 180) {
				img.Set(x, y, c)
			}
		}
	}

	// Accent color
	accent := color.RGBA{59, 130, 246, 255}  // blue-500
	if !isTUI {
		accent = color.RGBA{59, 130, 246, 255}
	}

	// Draw "T" letter - bold geometric
	drawT(img, size, accent)

	// Draw accent bar at bottom
	barY := size - 180
	barH := 40
	for y := barY; y < barY+barH; y++ {
		for x := 280; x < size-280; x++ {
			if isInRoundedRect(x-280, y-barY, size-560, barH, 20) {
				img.Set(x, y, accent)
			}
		}
	}

	// Draw small dots for "CODE" or "IDE"
	dotY := size - 110
	label := "CODE"
	if !isTUI {
		label = "IDE"
	}
	drawDots(img, size, dotY, accent, label)

	// Add subtle glow around T
	addGlow(img, size, accent)

	f, _ := os.Create(filename)
	defer f.Close()
	png.Encode(f, img)
}

func drawT(img *image.RGBA, size int, accent color.RGBA) {
	// Geometric "T" shape
	cx := size / 2

	// Top horizontal bar
	barW := 380
	barH := 80
	topY := 200
	for y := topY; y < topY+barH; y++ {
		for x := cx - barW/2; x < cx+barW/2; x++ {
			r := 20.0
			if isInRoundedRect(x-(cx-barW/2), y-topY, barW, barH, int(r)) {
				img.Set(x, y, accent)
			}
		}
	}

	// Vertical stem
	stemW := 90
	stemH := 420
	stemY := topY + barH - 10
	for y := stemY; y < stemY+stemH; y++ {
		for x := cx - stemW/2; x < cx+stemW/2; x++ {
			r := 20.0
			if isInRoundedRect(x-(cx-stemW/2), y-stemY, stemW, stemH, int(r)) {
				// Slight gradient on stem
				t := float64(y-stemY) / float64(stemH)
				c := lerpColor(accent, color.RGBA{accent.R, accent.G, accent.B, 180}, t*0.3)
				img.Set(x, y, c)
			}
		}
	}

	// Small ">" cursor symbol on the right
	cursorX := cx + 180
	cursorY := 460
	cursorSize := 60
	bright := color.RGBA{96, 165, 250, 255} // blue-400
	for i := 0; i < cursorSize; i++ {
		// Top diagonal
		py1 := cursorY + i
		px1 := cursorX + i/2
		for dx := 0; dx < 12; dx++ {
			img.Set(px1+dx, py1, bright)
		}
		// Bottom diagonal
		py2 := cursorY + cursorSize*2 - i
		px2 := cursorX + i/2
		for dx := 0; dx < 12; dx++ {
			img.Set(px2+dx, py2, bright)
		}
	}
}

func drawDots(img *image.RGBA, size, y int, accent color.RGBA, label string) {
	// Draw small text indicator dots
	cx := size / 2
	dotR := 8
	spacing := 28
	n := len(label)
	startX := cx - (n-1)*spacing/2

	dimAccent := color.RGBA{accent.R, accent.G, accent.B, 160}
	for i := 0; i < n; i++ {
		dx := startX + i*spacing
		for dy := y - dotR; dy <= y+dotR; dy++ {
			for ddx := dx - dotR; ddx <= dx+dotR; ddx++ {
				dist := math.Sqrt(float64((ddx-dx)*(ddx-dx) + (dy-y)*(dy-y)))
				if dist <= float64(dotR) {
					img.Set(ddx, dy, dimAccent)
				}
			}
		}
	}
}

func addGlow(img *image.RGBA, size int, accent color.RGBA) {
	// Subtle glow overlay at top
	for y := 0; y < size/3; y++ {
		t := 1.0 - float64(y)/float64(size/3)
		alpha := uint8(t * 8)
		glowColor := color.RGBA{accent.R, accent.G, accent.B, alpha}
		for x := 0; x < size; x++ {
			if isInRoundedRect(x, y, size, size, 180) {
				existing := img.RGBAAt(x, y)
				blended := blendOver(existing, glowColor)
				img.Set(x, y, blended)
			}
		}
	}
}

func isInRoundedRect(x, y, w, h, r int) bool {
	if x < 0 || y < 0 || x >= w || y >= h {
		return false
	}
	// Check corners
	if x < r && y < r {
		return (x-r)*(x-r)+(y-r)*(y-r) <= r*r
	}
	if x >= w-r && y < r {
		return (x-(w-r))*(x-(w-r))+(y-r)*(y-r) <= r*r
	}
	if x < r && y >= h-r {
		return (x-r)*(x-r)+(y-(h-r))*(y-(h-r)) <= r*r
	}
	if x >= w-r && y >= h-r {
		return (x-(w-r))*(x-(w-r))+(y-(h-r))*(y-(h-r)) <= r*r
	}
	return true
}

func lerpColor(a, b color.RGBA, t float64) color.RGBA {
	return color.RGBA{
		R: uint8(float64(a.R)*(1-t) + float64(b.R)*t),
		G: uint8(float64(a.G)*(1-t) + float64(b.G)*t),
		B: uint8(float64(a.B)*(1-t) + float64(b.B)*t),
		A: uint8(float64(a.A)*(1-t) + float64(b.A)*t),
	}
}

func blendOver(dst, src color.RGBA) color.RGBA {
	sa := float64(src.A) / 255
	return color.RGBA{
		R: uint8(float64(dst.R)*(1-sa) + float64(src.R)*sa),
		G: uint8(float64(dst.G)*(1-sa) + float64(src.G)*sa),
		B: uint8(float64(dst.B)*(1-sa) + float64(src.B)*sa),
		A: dst.A,
	}
}

// Ensure draw import is used
var _ = draw.Over
