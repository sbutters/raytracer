package main

import (
	"math"
	"math/rand"
	"sync"
)

type PixelTask struct {
	Row    int
	Height int
}

func rayColor(r Ray, world HittableList, depth int, rng *rand.Rand) Vec3 {
	if depth <= 0 {
		return Vec3{0, 0, 0}
	}

	var rec HitRecord
	if world.Hit(r, 0.001, math.MaxFloat64, &rec) {
		attenuation, scattered, ok := rec.Mat.Scatter(r, &rec, rng)
		if ok {
			return attenuation.Mul(rayColor(scattered, world, depth-1, rng))
		}
		return Vec3{0, 0, 0}
	}

	// Sky gradient
	unitDirection := r.Direction.Unit()
	t := 0.5 * (unitDirection.Y + 1.0)
	return Vec3{1, 1, 1}.Scale(1.0 - t).Add(Vec3{0.5, 0.7, 1.0}.Scale(t))
}

func renderRow(cam Camera, row, width, height, spp, maxDepth int, world HittableList, rng *rand.Rand, pixels *[][3]byte) {
	for i := 0; i < width; i++ {
		var pixelColor Vec3
		for s := 0; s < spp; s++ {
			ray := cam.GetRay(i, row, width, height, rng)
			color := rayColor(ray, world, maxDepth, rng)
			pixelColor = pixelColor.Add(color)
		}
		// Average samples
		pixelColor = pixelColor.Scale(1.0 / float64(spp))
		// Gamma correction (sqrt = gamma 2)
		pixelColor = Vec3{
			math.Sqrt(pixelColor.X),
			math.Sqrt(pixelColor.Y),
			math.Sqrt(pixelColor.Z),
		}
		// Clamp and convert to bytes
		b := pixelColor.Clamp().ToBytes()
		(*pixels)[row*width+i] = b
	}
}

func Render(width, height, spp, maxDepth, workers int, seed int64, world HittableList) [][3]byte {
	pixels := make([][3]byte, width*height)

	cam := NewCamera(
		Vec3{0, 1, 5.5},
		Vec3{0, 0.5, 0},
		Vec3{0, 1, 0},
		40,  // vfov
		float64(width)/float64(height),
		0,     // aperture
		5.5,   // focusDist = |lookFrom - lookAt|
		width, height,
	)

	// Create worker pool
	var wg sync.WaitGroup
	tasks := make(chan int, workers)

	for w := 0; w < workers; w++ {
		wg.Add(1)
		go func(workerID int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(seed + int64(workerID)))

			for row := range tasks {
				renderRow(cam, row, width, height, spp, maxDepth, world, rng, &pixels)
			}
		}(w)
	}

	// Feed tasks
	for row := 0; row < height; row++ {
		tasks <- row
	}
	close(tasks)
	wg.Wait()

	return pixels
}
