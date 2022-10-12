package math

type Vector2 struct {
	x int32
	y int32
}

func (v *Vector2) Add(v2 Vector2) {
	v.x += v2.x
	v.y += v2.y
}

func (v *Vector2) Sub(v2 Vector2) {
	v.x -= v2.x
	v.y -= v2.y
}
