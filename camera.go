package main

import (
	"math"
	"math/rand"
)

type Camera struct {
	Origin          Vec3
	LowerLeftCorner Vec3
	Horizontal      Vec3
	U, V            Vec3 // Basis vectors
	LensRadius      float64
}

func NewCamera(lookFrom, lookAt Vec3, vUp Vec3, vfov, aspectRatio, aperture, focusDist float64, width, height int) Camera {
	// vfov in degrees to radians
	u := vfov * math.Pi / 180.0
	halfViewportHeight := math.Tan(u / 2)
	focusLength := lookFrom.Sub(lookAt).Length()

	viewportHeight := 2.0 * halfViewportHeight * focusLength
	viewportWidth := viewportHeight * (float64(width) / float64(height))

	// Camera basis: w points FROM lookAt TO lookFrom (backwards from camera)
	w := lookFrom.Sub(lookAt).Unit()
	uVec := vUp.Cross(w).Unit()
	v := w.Cross(uVec)

	// Camera axes
	horizontal := uVec.Scale(viewportWidth)
	up := v.Scale(viewportHeight)

	// Lower left corner: lookFrom - half viewport - focusLength * w
	// Since w points backwards, -w*focusLength points forwards
	lowerLeftCorner := lookFrom.Sub(horizontal.Scale(0.5)).Sub(up.Scale(0.5)).Sub(w.Scale(focusLength))

	// Defocus disk
	lensRadius := aperture / 2.0

	return Camera{
		Origin:          lookFrom,
		LowerLeftCorner: lowerLeftCorner,
		Horizontal:      horizontal,
		U:               uVec,
		V:               up,
		LensRadius:      lensRadius,
	}
}

func (c Camera) GetRay(i, j int, width, height int, rng *rand.Rand) Ray {
	// Jittered pixel center
	sampleU := (float64(i) + rng.Float64()) / float64(width)
	sampleV := (float64(j) + rng.Float64()) / float64(height)

	// Ray through pixel
	rayOrigin := c.Origin
	rayDirection := c.LowerLeftCorner.Add(
		c.Horizontal.Scale(sampleU).Add(c.V.Scale(sampleV)),
	).Sub(c.Origin)

	// Defocus blur
	if c.LensRadius > 0 {
		pad := Vec3RandomInUnitDisk(rng).Scale(c.LensRadius)
		rayOrigin = c.Origin.Add(c.U.Scale(pad.X)).Add(c.V.Scale(pad.Y))
	}

	return Ray{rayOrigin, rayDirection}
}
