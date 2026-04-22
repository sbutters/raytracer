<p align="center">
  <img src="https://raw.githubusercontent.com/sbutters/raytracer/refs/heads/master/render.png" width="350" title="hover text">
</p>

# Raytracer Specification

## Overview

A CPU-based raytracing image renderer written in Go. Produces a PNG image of three spheres resting on a checkerboard ground plane, rendered in parallel across all available host processors.

## Language & Runtime

- **Language:** Go (1.21+)
- **Standard library only** for core rendering (`image/png`, `math`, `math/rand`, `sync`, `runtime`, `flag`, `os`)
- No third-party raytracing dependencies

## Command-Line Interface

```
raytracer [flags]
```

| Flag       | Type   | Default          | Description                          |
|------------|--------|------------------|--------------------------------------|
| `-width`   | int    | 1024             | Output image width in pixels         |
| `-height`  | int    | 1024             | Output image height in pixels        |
| `-out`     | string | `render.png`     | Output PNG file path                 |
| `-spp`     | int    | 100              | Samples per pixel (anti-aliasing)    |
| `-depth`   | int    | 50               | Max ray bounce depth                 |
| `-workers` | int    | `runtime.NumCPU()` | Number of parallel worker goroutines |
| `-seed`    | int64  | 1                | RNG seed for reproducibility         |

## Parallelism Model

- Determine worker count from `runtime.NumCPU()` (overridable via `-workers`).
- Divide the image into row-bands (scanline chunks) and feed them through a buffered channel to a pool of worker goroutines.
- Each worker owns its own `*rand.Rand` seeded from `-seed + workerID` to avoid contention on the global RNG.
- A `sync.WaitGroup` gates final PNG encoding until all workers have completed.
- Progress counter updated atomically (`atomic.AddInt64`) for optional stderr progress output.

## Package Layout

```
raytracer/
├── main.go              // flag parsing, orchestration, PNG encoding
├── vec3.go              // Vec3 type + math ops (Add, Sub, Dot, Cross, Unit, etc.)
├── ray.go               // Ray{Origin, Direction} + At(t)
├── hit.go               // HitRecord, Hittable interface, HittableList
├── sphere.go            // Sphere hittable
├── plane.go             // Infinite ground plane hittable (y = 0)
├── material.go          // Material interface + Lambertian, Metal, Dielectric, Checker
├── camera.go            // Camera with defocus blur + aspect ratio
├── scene.go             // Scene construction (spheres + ground)
└── render.go            // Worker pool, pixel sampling, color integration
```

## Core Types

### Vec3
```go
type Vec3 struct{ X, Y, Z float64 }
```
Methods: `Add`, `Sub`, `Mul`, `Scale`, `Dot`, `Cross`, `Length`, `LengthSquared`, `Unit`, `Reflect`, `Refract`, `NearZero`, plus helpers `RandomInUnitSphere`, `RandomUnitVector`, `RandomInUnitDisk`.

### Ray
```go
type Ray struct{ Origin, Direction Vec3 }
func (r Ray) At(t float64) Vec3
```

### Hittable
```go
type HitRecord struct {
    P         Vec3      // hit point
    Normal    Vec3      // outward normal
    T         float64   // ray parameter
    U, V      float64   // surface coords (for checker)
    FrontFace bool
    Mat       Material
}

type Hittable interface {
    Hit(r Ray, tMin, tMax float64, rec *HitRecord) bool
}
```

### Material
```go
type Material interface {
    Scatter(rIn Ray, rec *HitRecord, rng *rand.Rand) (attenuation Vec3, scattered Ray, ok bool)
}
```

Implementations:
- `Lambertian{Albedo Texture}` — matte diffuse; scatter in `rec.Normal + RandomUnitVector()`.
- `Metal{Albedo Vec3, Fuzz float64}` — reflective; fuzz=0 for mirror finish.
- `Dielectric{IR float64}` — glass; uses Snell's law + Schlick approximation for Fresnel.

### Texture
```go
type Texture interface { Value(u, v float64, p Vec3) Vec3 }
```
- `SolidColor{Color Vec3}`
- `Checker{Even, Odd Texture, Scale float64}` — evaluated via `sin(scale*x)*sin(scale*z) < 0` on the ground plane.

## Scene Specification

Camera:
- `lookFrom = (0, 1, 4)`
- `lookAt   = (0, 0.5, 0)`
- `vUp      = (0, 1, 0)`
- `vfov     = 40°`
- `aperture = 0.0` (sharp focus; bump to 0.1 for depth-of-field)
- `focusDist = |lookFrom - lookAt|`

Hittables:

| Object        | Center          | Radius | Material                                   |
|---------------|-----------------|--------|--------------------------------------------|
| Ground plane  | y = 0           | —      | Lambertian + Checker(red, white, scale=5)  |
| Left sphere   | (-1.1, 0.5, 0)  | 0.5    | Lambertian(white = {0.9, 0.9, 0.9})        |
| Middle sphere | ( 0.0, 0.5, 0)  | 0.5    | Dielectric(IR = 1.5)                       |
| Right sphere  | ( 1.1, 0.5, 0)  | 0.5    | Metal(albedo = {0.25, 0.25, 0.25}, fuzz=0) |

Sky / background: linear gradient from white `(1, 1, 1)` to pale blue `(0.5, 0.7, 1.0)` along ray direction Y.

## Rendering Pipeline

For each pixel `(i, j)`:
1. Take `spp` samples with jittered offsets `(u, v) = ((i + rand)/W, (j + rand)/H)`.
2. Generate ray through camera.
3. Call `rayColor(r, world, depth)`:
   - If `depth <= 0` → return `Vec3{0,0,0}`.
   - If no hit → return sky gradient.
   - If material scatters → return `attenuation * rayColor(scattered, world, depth-1)`.
   - Else → return black.
4. Average the samples, apply gamma-2 correction (`sqrt`), clamp to `[0, 0.999]`, convert to 8-bit sRGB.

## Depth Selection Rationale

Dielectrics (glass) are the depth-hungry case: each hit can split into a reflected + refracted path and the middle sphere sits directly behind the camera's line of sight. Empirical behavior:

| Depth | Behavior                                                |
|-------|---------------------------------------------------------|
| 10    | Glass sphere shows black cores where refraction terminates |
| 25    | Mostly correct; faint darkening inside glass            |
| **50** | **Default — glass fully resolves; no visible clipping** |
| 100   | No perceptible improvement; ~1.6x render cost           |

Default to **50**; expose via `-depth` flag for tuning.

## Output

- 8-bit RGBA PNG via `image/png.Encode`.
- Origin `(0,0)` at top-left (invert Y from camera space when writing pixels).
- Write atomically: render to `<out>.tmp`, then `os.Rename` to `<out>`.

## Performance Targets (reference, 8-core machine)

| Resolution | SPP | Depth | Expected Time |
|------------|-----|-------|---------------|
| 400x400    | 50  | 50    | ~2 s          |
| 1024x1024  | 100 | 50    | ~45 s         |
| 1920x1080  | 500 | 50    | ~8 min        |

## Implementation Order

1. `Vec3`, `Ray` + unit tests for dot/cross/reflect/refract.
2. `Sphere.Hit`, `HittableList`.
3. `Camera` + single-threaded render of a solid-color sphere → sanity PNG.
4. `Lambertian` + sky gradient.
5. `Metal` with fuzz.
6. `Dielectric` with Schlick Fresnel.
7. Ground plane + `Checker` texture.
8. Worker-pool parallelism + CLI flags.
9. Gamma correction + tone mapping polish.
10. Final scene composition matching spec table above.
