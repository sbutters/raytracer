package main

import "math"

type Sphere struct {
	Center Vec3
	Radius float64
	Mat    Material
}

func NewSphere(center Vec3, radius float64, mat Material) Sphere {
	return Sphere{center, radius, mat}
}

func (s Sphere) Hit(r Ray, tMin, tMax float64, rec *HitRecord) bool {
	oc := r.Origin.Sub(s.Center)
	a := r.Direction.LengthSquared()
	halfB := oc.Dot(r.Direction)
	c := oc.LengthSquared() - s.Radius*s.Radius
	discriminant := halfB*halfB - a*c

	if discriminant < 0 {
		return false
	}

	sqrtDisc := math.Sqrt(discriminant)

	// Use the smaller positive root
	root := (-halfB - sqrtDisc) / a
	if root <= tMin || root >= tMax {
		root = (-halfB + sqrtDisc) / a
		if root <= tMin || root >= tMax {
			return false
		}
	}

	rec.T = root
	rec.P = r.At(root)

	// Compute normal
	outwardNormal := rec.P.Sub(s.Center).Scale(1.0 / s.Radius)
	rec.SetFaceNormal(r, outwardNormal)

	// Compute UV coordinates for texture mapping
	// Project sphere point onto UV plane
	theta := math.Acos(-outwardNormal.Y)
	phi := math.Atan2(-outwardNormal.Z, outwardNormal.X) + math.Pi
	rec.U = phi / (2 * math.Pi)
	rec.V = theta / math.Pi

	rec.Mat = s.Mat
	return true
}
