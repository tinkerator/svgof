// Program drl2svg converts Kicad generated drl files into mm accurate
// SVG files that can be CNCd on a Snapmaker 2.0 A350T CNC.
//
// Given a --bit-size, all holes can be drilled as concentric circles
// of 45% of the bit size increments of the radius. We generate a SVG
// with this pattern to be sure each hole disintegrates, and doesn't
// create chunks that interrupt the smooth operation of the CNC.
package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"log"
	"os"

	"zappem.net/pub/graphics/svgof"
)

var (
	drl     = flag.String("drl", "/dev/stdin", "drill file")
	bitSize = flag.Float64("bit-size", .7, "CNC bit diameter, in mm")
	debug   = flag.Bool("debug", false, "log more debugging information")
	dest    = flag.String("dest", "", "output SVG filename (or, /dev/stdout)")
)

type Hole struct {
	Radius float64
	Cx     float64
	Cy     float64
}

func main() {
	flag.Parse()

	f, err := os.Open(*drl)
	if err != nil {
		log.Fatalf("unable to read %q: %v", *drl, err)
	}
	defer f.Close()

	tools := make(map[string]float64)
	var holes []Hole

	sc := bufio.NewScanner(f)

	var toolRadius, maxToolDiameter float64
	factor := 0.0
	var leftEdge, rightEdge, topEdge, bottomEdge float64

	// Extract hole information out of the drl file.

	for sc.Scan() {
		var tool int
		var param1, param2 float64
		line := sc.Text()

		if line == "METRIC" {
			factor = 1.0
			continue
		}
		if line == "INCH" {
			factor = 25.4
			continue
		}
		if n, err := fmt.Sscanf(line, "T%dC%f", &tool, &param1); err == nil && n == 2 {
			if param1 < *bitSize {
				log.Fatalf("unable to handle tool diameter %q < %f mm", line, *bitSize)
			}
			d := param1 * factor
			if len(tools) == 0 || d > maxToolDiameter {
				maxToolDiameter = d
			}
			tools[fmt.Sprint("T", tool)] = d
			continue
		}
		if d, ok := tools[line]; ok {
			toolRadius = (d - *bitSize) * 0.5
			continue
		}
		if n, err := fmt.Sscanf(line, "X%fY%f", &param1, &param2); err == nil && n == 2 {
			h := Hole{
				Radius: toolRadius,
				Cx:     param1 * factor,
				Cy:     param2 * factor,
			}
			if left := h.Cx - toolRadius; len(holes) == 0 || left < leftEdge {
				leftEdge = left
			}
			if right := h.Cx + toolRadius; len(holes) == 0 || right > rightEdge {
				rightEdge = right
			}
			if top := h.Cy - toolRadius; len(holes) == 0 || top < topEdge {
				topEdge = top
			}
			if bottom := h.Cy + toolRadius; len(holes) == 0 || bottom > bottomEdge {
				bottomEdge = bottom
			}
			holes = append(holes, h)
		} else {
			if *debug {
				log.Printf("ignored: %q", line)
			}
			continue
		}
	}

	if *debug {
		log.Printf("tools loaded: %#v", tools)
	}

	leftEdge -= maxToolDiameter
	rightEdge += maxToolDiameter
	topEdge -= maxToolDiameter
	bottomEdge += maxToolDiameter

	var out io.Writer

	if *dest != "" {
		wr, err := os.Create(*dest)
		if err != nil {
			log.Fatalf("failed to create %q: %v", *dest, err)
		}
		defer wr.Close()
		out = wr
	} else {
		out = os.Stdout
	}

	canvas := svgof.New(out)
	canvas.Decimals = 3

	// We declare "mm" here to be explicit about the units.
	canvas.StartviewUnit(rightEdge-leftEdge, bottomEdge-topEdge, "mm", leftEdge, topEdge, rightEdge-leftEdge, bottomEdge-topEdge)
	for _, h := range holes {
		var radii []float64
		for dr := h.Radius; dr > 0; dr -= 0.45 * *bitSize {
			radii = append(radii, dr)
		}
		for i := len(radii) - 1; i >= 0; i-- {
			canvas.Circle(h.Cx, h.Cy, radii[i])
		}
	}
	canvas.End()
}
