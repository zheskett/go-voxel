package tensor

import (
	"github.com/chewxy/math32"
)

// A 2D Vector
type Vector2 struct {
	X, Y float32 // X coordinate
}

// A 3D Vector
type Vector3 struct {
	X, Y, Z float32 // X coordinate
}

// Start Vector2 Functions

// Returns the Elements of the vector
func (v Vector2) Elms() (float32, float32) {
	return v.X, v.Y
}

// Returns the length of the vector
func (v Vector2) Len() float32 {
	return math32.Hypot(v.X, v.Y)
}

// Returns the squared length of the vector
func (v Vector2) LenSqr() float32 {
	return v.X*v.X + v.Y*v.Y
}

// Returns the normalized vector
func (v Vector2) Normalized() Vector2 {
	invLen := 1.0 / v.Len()
	return Vector2{v.X * invLen, v.Y * invLen}
}

// Returns the sum of two vectors
func (v1 Vector2) Add(v2 Vector2) Vector2 {
	return Vector2{v1.X + v2.X, v1.Y + v2.Y}
}

// Returns the difference of two vectors
func (v1 Vector2) Sub(v2 Vector2) Vector2 {
	return Vector2{v1.X - v2.X, v1.Y - v2.Y}
}

// Returns the product of a vector and a scalar
func (v1 Vector2) Mul(c float32) Vector2 {
	return Vector2{v1.X * c, v1.Y * c}
}

// Returns the quotient of a vector and a scalar
func (v1 Vector2) Div(c float32) Vector2 {
	invC := 1.0 / c
	return Vector2{v1.X * invC, v1.Y * invC}
}

// Returns the dot product of two vectors
func (v1 Vector2) Dot(v2 Vector2) float32 {
	return v1.X*v2.X + v1.Y*v2.Y
}

// Returns the cross product of two vectors
func (v1 Vector2) Cross(v2 Vector2) Vector2 {
	return Vector2{v1.X*v2.Y - v1.Y*v2.X, v1.Y*v2.X - v1.X*v2.Y}
}

// End Vector2 Functions

// Start Vector3 Functions

// Returns the Elements of the vector
func (v Vector3) Elms() (float32, float32, float32) {
	return v.X, v.Y, v.Z
}

// Returns the length of the vector
func (v Vector3) Len() float32 {
	return math32.Sqrt(v.X*v.X + v.Y*v.Y + v.Z*v.Z)
}

// Returns the squared length of the vector
func (v Vector3) LenSqr() float32 {
	return v.X*v.X + v.Y*v.Y + v.Z*v.Z
}

// Returns the normalized vector
func (v Vector3) Normalized() Vector3 {
	invLen := 1.0 / v.Len()
	return Vector3{v.X * invLen, v.Y * invLen, v.Z * invLen}
}

// Returns the sum of two vectors
func (v1 Vector3) Add(v2 Vector3) Vector3 {
	return Vector3{v1.X + v2.X, v1.Y + v2.Y, v1.Z + v2.Z}
}

// Returns the difference of two vectors
func (v1 Vector3) Sub(v2 Vector3) Vector3 {
	return Vector3{v1.X - v2.X, v1.Y - v2.Y, v1.Z - v2.Z}
}

// Returns the product of a vector and a scalar
func (v1 Vector3) Mul(c float32) Vector3 {
	return Vector3{v1.X * c, v1.Y * c, v1.Z * c}
}

// Returns the quotient of a vector and a scalar
func (v1 Vector3) Div(c float32) Vector3 {
	invC := 1.0 / c
	return Vector3{v1.X * invC, v1.Y * invC, v1.Z * invC}
}

// Returns the dot product of two vectors
func (v1 Vector3) Dot(v2 Vector3) float32 {
	return v1.X*v2.X + v1.Y*v2.Y + v1.Z*v2.Z
}

// Returns the cross product of two vectors
func (v1 Vector3) Cross(v2 Vector3) Vector3 {
	return Vector3{
		v1.Y*v2.Z - v1.Z*v2.Y,
		v1.Z*v2.X - v1.X*v2.Z,
		v1.X*v2.Y - v1.Y*v2.X,
	}
}

// End Vector3 Functions
