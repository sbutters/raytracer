package main

import (
	"math"
)

// Plane is an infinite ground plane at y = 0
type Plane struct {
	Mat Material
}

func NewPlane(mat Material) Plane {
	return Plane{mat}
}

func (p Plane) Hit(r Ray, tMin, tMax float64, rec *HitRecord) bool {
	// Plane equation: y = 0, normal = (0, 1, 0)
	// Intersection: r.origin.y + r.direction.y * t = 0
	// t = -r.origin.y / r.direction.y
	if math.Abs(r.Direction.Y) < 1e-8 {
		return false
	}

	t := -r.Origin.Y / r.Direction.Y
	if t <= tMin || t >= tMax {
		return false
	}

	rec.T = t
	rec.P = r.At(t)
	rec.Normal = Vec3{0, 1, 0}
	rec.SetFaceNormal(r, rec.Normal)

	// UV coordinates from X and Z positions
	rec.U = rec.P.X
	rec.V = rec.P.Z

	rec.Mat = p.Mat
	return true
}
