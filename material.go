package main

import (
	"math"
	"math/rand"
)

type Material interface {
	Scatter(rIn Ray, rec *HitRecord, rng *rand.Rand) (attenuation Vec3, scattered Ray, ok bool)
}

type Texture interface {
	Value(u, v float64, p Vec3) Vec3
}

// SolidColor is a texture that returns a constant color
type SolidColor struct {
	Color Vec3
}

func (s SolidColor) Value(u, v float64, p Vec3) Vec3 {
	return s.Color
}

// CheckerTexture alternates between two colors based on grid position
type CheckerTexture struct {
	Even Texture
	Odd  Texture
	Scale float64
}

func NewCheckerTexture(scale float64, even Texture, odd Texture) CheckerTexture {
	return CheckerTexture{Even: even, Odd: odd, Scale: scale}
}

func (ct CheckerTexture) Value(u, v float64, p Vec3) Vec3 {
	// Checker pattern using sin as specified
	scaledX := p.X * ct.Scale
	scaledZ := p.Z * ct.Scale
	if math.Sin(scaledX)*math.Sin(scaledZ) < 0 {
		return ct.Odd.Value(u, v, p)
	}
	return ct.Even.Value(u, v, p)
}

// Lambertian has a non-metallic, diffuse material
type Lambertian struct {
	Albedo Texture
}

func NewLambertian(color Vec3) Lambertian {
	return Lambertian{Albedo: SolidColor{Color: color}}
}

func (la Lambertian) Scatter(rIn Ray, rec *HitRecord, rng *rand.Rand) (Vec3, Ray, bool) {
	scatterDirection := rec.Normal.Add(Vec3RandomUnitVector(rng))

	// Catch degenerate scatter direction (when random vector is nearly opposite normal)
	if scatterDirection.NearZero() {
		scatterDirection = rec.Normal
	}

	scattered := Ray{rec.P, scatterDirection}
	attenuation := la.Albedo.Value(rec.U, rec.V, rec.P)
	return attenuation, scattered, true
}

// Metal reflects light in a mirror direction with some fuzz
type Metal struct {
	Albedo Vec3
	Fuzz   float64
}

func NewMetal(albedo Vec3, fuzz float64) Metal {
	if fuzz < 1 {
		return Metal{Albedo: albedo, Fuzz: fuzz}
	}
	return Metal{Albedo: albedo, Fuzz: 1}
}

func (m Metal) Scatter(rIn Ray, rec *HitRecord, rng *rand.Rand) (Vec3, Ray, bool) {
	reflected := rIn.Direction.Unit().Reflect(rec.Normal)
	scattered := Ray{rec.P, reflected.Add(Vec3RandomInUnitSphere(rng).Scale(m.Fuzz))}

	// Only scatter if the reflected direction points outward
	if scattered.Direction.Dot(rec.Normal) > 0 {
		return m.Albedo, scattered, true
	}
	return Vec3{}, Ray{}, false
}

// Dielectric is a transparent material (glass)
type Dielectric struct {
	IR float64 // Index of refraction
}

func NewDielectric(ir float64) Dielectric {
	return Dielectric{IR: ir}
}

func (d Dielectric) Scatter(rIn Ray, rec *HitRecord, rng *rand.Rand) (Vec3, Ray, bool) {
	attenuation := Vec3{1, 1, 1}
	refractionRatio := d.IR
	if rec.FrontFace {
		refractionRatio = 1.0 / d.IR
	}

	// rec.Normal always points against the incoming ray, so negate to get cosTheta in [0,1]
	cosTheta := math.Min(-rIn.Direction.Unit().Dot(rec.Normal), 1.0)
	sinTheta := math.Sqrt(1.0 - cosTheta*cosTheta)

	// Total internal reflection
	cannotRefract := refractionRatio*sinTheta > 1.0

	var direction Vec3
	if cannotRefract || schlick(cosTheta, d.IR) > rng.Float64() {
		// Reflect
		direction = rIn.Direction.Unit().Reflect(rec.Normal)
	} else {
		// Refract — rec.Normal already faces against the incoming ray, no flip needed
		direction = rIn.Direction.Unit().Refract(rec.Normal, refractionRatio)
	}

	scattered := Ray{rec.P, direction}
	return attenuation, scattered, true
}

// Schlick computes the Fresnel reflectance for dielectric interfaces
func schlick(cosine float64, refractionIndex float64) float64 {
	r0 := (1 - refractionIndex) / (1 + refractionIndex)
	r0 = r0 * r0
	return r0 + (1-r0)*math.Pow(1-cosine, 5)
}
