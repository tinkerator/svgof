// Program outer explores how to inflate a line by some distance.
// This has been used to visualize improvements to the
// [zappem.net/pub/math/polygon.Shapes.Inflate] function.
package main

import (
	"flag"
	"log"
	"math"
	"os"

	"zappem.net/pub/graphics/svgof"
)

var (
	dest  = flag.String("dest", "", "destination filename")
	grow  = flag.Float64("grow", 10.0, "how much to grow the line")
	alpha = flag.Float64("alpha", 30.0, "subtended angle at 1st corner (deg)")
	beta  = flag.Float64("beta", 30.0, "subtended angle at 2nd corner (deg)")
	mid   = flag.Float64("mid", 40.0, "length of between line (mm)")
	ext   = flag.Float64("ext", 50.0, "length of exterior lines (mm)")
)

func main() {
	flag.Parse()

	out := os.Stdout
	var err error

	if *dest != "" {
		out, err = os.Create(*dest)
		if err != nil {
			log.Fatalf("Unable to write %q: %v", *dest, err)
		}
		defer out.Close()
	}

	ang := func(theta float64) float64 {
		return theta / 180.0 * math.Pi
	}
	min := func(xs ...float64) float64 {
		var x0 float64
		for i, x := range xs {
			if i == 0 || x < x0 {
				x0 = x
			}
		}
		return x0
	}
	max := func(xs ...float64) float64 {
		var x0 float64
		for i, x := range xs {
			if i == 0 || x > x0 {
				x0 = x
			}
		}
		return x0
	}

	cA, sA := math.Cos(ang(*alpha)), math.Sin(ang(*alpha))
	cAB, sAB := math.Cos(ang(*alpha+*beta)), math.Sin(ang(*alpha+*beta))

	x0, y0 := 0.0, 0.0
	x1, y1 := *ext, 0.0
	x2, y2 := *ext+*mid*cA, -*mid*sA
	x3, y3 := x2+*ext*cAB, y2-*ext*sAB

	minX := min(x0, x1, x2, x3) - 2**grow
	minY := min(y0, y1, y2, y3) - 2**grow
	maxX := max(x0, x1, x2, x3) + 2**grow
	maxY := max(y0, y1, y2, y3) + 2**grow

	canvas := svgof.New(out)
	canvas.Decimals = 3

	canvas.StartviewUnit(maxX-minX, maxY-minY, "mm", minX, minY, maxX-minX, maxY-minY)
	canvas.Rect(minX, minY, maxX-minX, maxY-minY, `fill="white"`)
	canvas.Polyline([]float64{x0, x1, x2, x3}, []float64{y0, y1, y2, y3}, `fill="none" stroke="blue" stroke-width=".2"`)

	style := `fill="none" stroke="black" stroke-width=".1"`
	canvas.Circle(0.5*(x0+x1), 0.5*(y0+y1), *grow, style)
	canvas.Circle(x1, y1, *grow, style+` stroke-dasharray="2, 1"`)
	canvas.Circle(0.5*(x1+x2), 0.5*(y1+y2), *grow, style)
	canvas.Circle(x2, y2, *grow, style+` stroke-dasharray="2, 1"`)
	canvas.Circle(0.5*(x2+x3), 0.5*(y2+y3), *grow, style)

	u1X, u1Y := (x1-x0) / *ext, (y1-y0) / *ext
	p1X, p1Y := u1Y, -u1X
	u2X, u2Y := (x2-x1) / *mid, (y2-y1) / *mid
	p2X, p2Y := u2Y, -u2X
	u3X, u3Y := (x3-x2) / *ext, (y3-y2) / *ext
	p3X, p3Y := u3Y, -u3X

	style = `fill="none" stroke="red" stroke-width=".1"`

	// inside
	canvas.Line(0.5*(x0+x1+(*ext+*grow*2)*u1X)+*grow*p1X, 0.5*(y0+y1+(*ext+*grow*2)*u1Y)+*grow*p1Y, 0.5*(x0+x1-(*ext+*grow*2)*u1X)+*grow*p1X, 0.5*(y0+y1-(*ext+*grow*2)*u1Y)+*grow*p1Y, style)
	canvas.Line(0.5*(x1+x2+(*ext+*grow*2)*u2X)+*grow*p2X, 0.5*(y1+y2+(*ext+*grow*2)*u2Y)+*grow*p2Y, 0.5*(x1+x2-(*ext+*grow*2)*u2X)+*grow*p2X, 0.5*(y1+y2-(*ext+*grow*2)*u2Y)+*grow*p2Y, style)
	canvas.Line(0.5*(x2+x3+(*ext+*grow*2)*u3X)+*grow*p3X, 0.5*(y2+y3+(*ext+*grow*2)*u3Y)+*grow*p3Y, 0.5*(x2+x3-(*ext+*grow*2)*u3X)+*grow*p3X, 0.5*(y2+y3-(*ext+*grow*2)*u3Y)+*grow*p3Y, style)
	// outside
	canvas.Line(0.5*(x0+x1+(*ext+*grow*2)*u1X)-*grow*p1X, 0.5*(y0+y1+(*ext+*grow*2)*u1Y)-*grow*p1Y, 0.5*(x0+x1-(*ext+*grow*2)*u1X)-*grow*p1X, 0.5*(y0+y1-(*ext+*grow*2)*u1Y)-*grow*p1Y, style)
	canvas.Line(0.5*(x1+x2+(*ext+*grow*2)*u2X)-*grow*p2X, 0.5*(y1+y2+(*ext+*grow*2)*u2Y)-*grow*p2Y, 0.5*(x1+x2-(*ext+*grow*2)*u2X)-*grow*p2X, 0.5*(y1+y2-(*ext+*grow*2)*u2Y)-*grow*p2Y, style)
	canvas.Line(0.5*(x2+x3+(*ext+*grow*2)*u3X)-*grow*p3X, 0.5*(y2+y3+(*ext+*grow*2)*u3Y)-*grow*p3Y, 0.5*(x2+x3-(*ext+*grow*2)*u3X)-*grow*p3X, 0.5*(y2+y3-(*ext+*grow*2)*u3Y)-*grow*p3Y, style)

	canvas.End()
}
