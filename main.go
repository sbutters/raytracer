package main

import (
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"os"
	"runtime"
)

func main() {
	width := flag.Int("width", 1024, "Output image width in pixels")
	height := flag.Int("height", 1024, "Output image height in pixels")
	out := flag.String("out", "render.png", "Output PNG file path")
	spp := flag.Int("spp", 100, "Samples per pixel (anti-aliasing)")
	depth := flag.Int("depth", 50, "Max ray bounce depth")
	workers := flag.Int("workers", runtime.NumCPU(), "Number of parallel worker goroutines")
	seed := flag.Int64("seed", 1, "RNG seed for reproducibility")
	flag.Parse()

	fmt.Fprintf(os.Stderr, "Raytracer: %dx%d, %d spp, depth %d, %d workers, seed %d\n",
		*width, *height, *spp, *depth, *workers, *seed)

	world := BuildScene()

	fmt.Fprintf(os.Stderr, "Rendering...\n")
	pixels := Render(*width, *height, *spp, *depth, *workers, *seed, world)
	fmt.Fprintf(os.Stderr, "Encoding PNG...\n")

	// Create image
	img := image.NewRGBA(image.Rect(0, 0, *width, *height))
	for y := 0; y < *height; y++ {
		for x := 0; x < *width; x++ {
			idx := y*(*width) + x
			b := pixels[idx]
			img.Set(x, *height-1-y, color.RGBA{R: b[0], G: b[1], B: b[2], A: 255})
		}
	}

	// Write atomically: render to .tmp, then rename
	tmpFile := *out + ".tmp"
	f, err := os.Create(tmpFile)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error creating temp file: %v\n", err)
		os.Exit(1)
	}
	err = png.Encode(f, img)
	if err != nil {
		f.Close()
		os.Remove(tmpFile)
		fmt.Fprintf(os.Stderr, "Error encoding PNG: %v\n", err)
		os.Exit(1)
	}
	err = f.Close()
	if err != nil {
		os.Remove(tmpFile)
		fmt.Fprintf(os.Stderr, "Error closing file: %v\n", err)
		os.Exit(1)
	}

	err = os.Rename(tmpFile, *out)
	if err != nil {
		os.Remove(tmpFile)
		fmt.Fprintf(os.Stderr, "Error renaming file: %v\n", err)
		os.Exit(1)
	}

	fmt.Fprintf(os.Stderr, "Done! Output written to %s\n", *out)
}
