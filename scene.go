package main

func BuildScene() HittableList {
	var world HittableList

	// Ground plane with checker texture
	evenColor := Vec3{0.5, 0.5, 0.5}
	oddColor := Vec3{1.0, 1.0, 1.0}
	checker := NewCheckerTexture(5, SolidColor{Color: evenColor}, SolidColor{Color: oddColor})
	world.Add(NewPlane(Lambertian{Albedo: checker}))

	// Left sphere: Lambertian white
	world.Add(NewSphere(Vec3{-1.1, 0.5, 0.0}, 0.5, NewLambertian(Vec3{0.9, 0.9, 0.9})))

	// Middle sphere: Dielectric (glass)
	world.Add(NewSphere(Vec3{0.0, 0.5, 0.0}, 0.5, NewDielectric(1.5)))

	// Right sphere: Metal
	world.Add(NewSphere(Vec3{1.1, 0.5, 0.0}, 0.5, NewMetal(Vec3{0.25, 0.25, 0.25}, 0)))

	return world
}
