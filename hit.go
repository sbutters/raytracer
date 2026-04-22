package main

type HitRecord struct {
	P         Vec3
	Normal    Vec3
	T         float64
	U, V      float64
	FrontFace bool
	Mat       Material
}

func (h *HitRecord) SetFaceNormal(r Ray, outwardNormal Vec3) {
	h.FrontFace = r.Direction.Dot(outwardNormal) < 0
	if h.FrontFace {
		h.Normal = outwardNormal
	} else {
		h.Normal = outwardNormal.Scale(-1)
	}
}

type Hittable interface {
	Hit(r Ray, tMin, tMax float64, rec *HitRecord) bool
}

type HittableList struct {
	Objects []Hittable
}

func NewHittableList() HittableList {
	return HittableList{}
}

func (hl *HittableList) Add(h Hittable) {
	hl.Objects = append(hl.Objects, h)
}

func (hl HittableList) Hit(r Ray, tMin, tMax float64, rec *HitRecord) bool {
	hitAnything := false
	closest := tMax
	tempRec := HitRecord{}

	for _, obj := range hl.Objects {
		if obj.Hit(r, tMin, closest, &tempRec) {
			hitAnything = true
			closest = tempRec.T
			*rec = tempRec
		}
	}

	return hitAnything
}
