package main

import (
	"math"
	"math/rand"
)

func (v Vec3) ToBytes() [3]byte {
	return [3]byte{
		byte(math.Floor(v.X * 255.999)),
		byte(math.Floor(v.Y * 255.999)),
		byte(math.Floor(v.Z * 255.999)),
	}
}

func Vec3RandomUnitVector(rng *rand.Rand) Vec3 {
	for {
		p := Vec3New(
			rng.Float64()*2-1,
			rng.Float64()*2-1,
			rng.Float64()*2-1,
		)
		l2 := p.LengthSquared()
		if l2 > 0 && l2 < 1 {
			return p.Unit()
		}
	}
}

func Vec3RandomInUnitSphere(rng *rand.Rand) Vec3 {
	for {
		p := Vec3New(
			rng.Float64()*2-1,
			rng.Float64()*2-1,
			rng.Float64()*2-1,
		)
		if p.LengthSquared() < 1 {
			return p
		}
	}
}

func Vec3RandomInUnitDisk(rng *rand.Rand) Vec3 {
	for {
		p := Vec3New(
			rng.Float64()*2-1,
			0,
			rng.Float64()*2-1,
		)
		if p.LengthSquared() < 1 {
			return p
		}
	}
}

type Vec3 struct {
	X, Y, Z float64
}

func Vec3New(x, y, z float64) Vec3 {
	return Vec3{x, y, z}
}

func Vec3FromSlice(s []float64) Vec3 {
	return Vec3{s[0], s[1], s[2]}
}

func (v Vec3) Add(w Vec3) Vec3 {
	return Vec3{v.X + w.X, v.Y + w.Y, v.Z + w.Z}
}

func (v Vec3) Sub(w Vec3) Vec3 {
	return Vec3{v.X - w.X, v.Y - w.Y, v.Z - w.Z}
}

func (v Vec3) Mul(w Vec3) Vec3 {
	return Vec3{v.X * w.X, v.Y * w.Y, v.Z * w.Z}
}

func (v Vec3) Scale(s float64) Vec3 {
	return Vec3{v.X * s, v.Y * s, v.Z * s}
}

func (v Vec3) Dot(w Vec3) float64 {
	return v.X*w.X + v.Y*w.Y + v.Z*w.Z
}

func (v Vec3) Cross(w Vec3) Vec3 {
	return Vec3{
		v.Y*w.Z - v.Z*w.Y,
		v.Z*w.X - v.X*w.Z,
		v.X*w.Y - v.Y*w.X,
	}
}

func (v Vec3) Length() float64 {
	return math.Sqrt(v.LengthSquared())
}

func (v Vec3) LengthSquared() float64 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

func (v Vec3) Unit() Vec3 {
	return v.Scale(1.0 / v.Length())
}

func (v Vec3) Reflect(n Vec3) Vec3 {
	return v.Sub(n.Scale(2.0 * v.Dot(n)))
}

func (v Vec3) Refract(nv Vec3, etaiOverEta float64) Vec3 {
	u := v.Unit()
	dt := u.Dot(nv)
	discriminant := 1.0 - etaiOverEta*etaiOverEta*(1.0-dt*dt)
	if discriminant > 0 {
		return u.Scale(etaiOverEta).Sub(nv.Scale(etaiOverEta*dt + math.Sqrt(discriminant)))
	}
	return Vec3{}
}

func (v Vec3) NearZero() bool {
	s := 1e-8
	return math.Abs(v.X) < s && math.Abs(v.Y) < s && math.Abs(v.Z) < s
}

func (v Vec3) Clamp() Vec3 {
	r := v.X
	g := v.Y
	b := v.Z
	if r < 0 {
		r = 0
	}
	if r > 0.999 {
		r = 0.999
	}
	if g < 0 {
		g = 0
	}
	if g > 0.999 {
		g = 0.999
	}
	if b < 0 {
		b = 0
	}
	if b > 0.999 {
		b = 0.999
	}
	return Vec3{r, g, b}
}
