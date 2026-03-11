package icon

import (
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
)

// Normal returns a dark gray circle icon (no new notifications).
func Normal() []byte {
	return render(color.RGBA{0x3a, 0x3a, 0x3a, 0xff}, false)
}

// Alert returns an orange circle icon (new notifications pending).
func Alert() []byte {
	return render(color.RGBA{0xe3, 0x6b, 0x00, 0xff}, true)
}

func render(bg color.RGBA, dot bool) []byte {
	const size = 32
	img := image.NewRGBA(image.Rect(0, 0, size, size))

	cx, cy := float64(size)/2-0.5, float64(size)/2-0.5
	outerR := float64(size)/2 - 1
	innerR := outerR * 0.55

	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			d := math.Sqrt(dx*dx + dy*dy)
			if d <= outerR {
				img.Set(x, y, bg)
			}
		}
	}

	// Draw a simple bell silhouette in white inside the circle.
	// Bell body (arc approximation using filled ellipse)
	bellColor := color.RGBA{0xff, 0xff, 0xff, 0xff}
	for y := 8; y <= 22; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			// Trapezoid-like bell body
			halfW := 4.0 + float64(y-8)*0.45
			if y <= 20 && math.Abs(dx) <= halfW {
				img.Set(x, y, bellColor)
			}
		}
	}
	// Bell top arc
	for y := 6; y <= 14; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - 14
			rx, ry := 5.0, 9.0
			if (dx*dx)/(rx*rx)+(dy*dy)/(ry*ry) <= 1.0 {
				img.Set(x, y, bellColor)
			}
		}
	}
	// Clapper dot
	clapperY := 23.0
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dcy := float64(y) - clapperY
			if math.Sqrt(dx*dx+dcy*dcy) <= 1.5 {
				img.Set(x, y, bellColor)
			}
		}
	}
	// Handle notch at top
	for y := 5; y <= 7; y++ {
		for x := 14; x <= 17; x++ {
			img.Set(x, y, bellColor)
		}
	}

	// Red dot on top-right when alert
	if dot {
		dotColor := color.RGBA{0xff, 0x30, 0x30, 0xff}
		dotCX, dotCY := cx+8.0, cy-8.0
		for y := 0; y < size; y++ {
			for x := 0; x < size; x++ {
				dx := float64(x) - dotCX
				dy := float64(y) - dotCY
				if math.Sqrt(dx*dx+dy*dy) <= 4.0 {
					img.Set(x, y, dotColor)
				}
			}
		}
	}

	// Mask to circle again (clean edges)
	for y := 0; y < size; y++ {
		for x := 0; x < size; x++ {
			dx := float64(x) - cx
			dy := float64(y) - cy
			if math.Sqrt(dx*dx+dy*dy) > outerR {
				img.Set(x, y, color.RGBA{0, 0, 0, 0})
			}
		}
	}

	_ = innerR
	var buf bytes.Buffer
	_ = png.Encode(&buf, img)
	return buf.Bytes()
}
