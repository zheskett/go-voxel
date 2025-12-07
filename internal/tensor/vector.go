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

// Creates a 2D Vector from components
func Vec2(x, y float32) Vector2 {
	return Vector2{X: x, Y: y}
}

// Creates the 2D zero vector
func Vec2Zero() Vector2 {
	return Vector2{X: 0, Y: 0}
}

// Creates a 2D vector with similar components
func Vec2Splat(c float32) Vector2 {
	return Vector2{X: c, Y: c}
}

// Creates an x unit vector
func Vec2X() Vector2 {
	return Vector2{X: 1, Y: 0}
}

// Creates an y unit vector
func Vec2Y() Vector2 {
	return Vector2{X: 0, Y: 1}
}

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

// Returns either the normalized vector or the zero vector if the length is too small
func (v Vector2) NormalizedOrZero() Vector2 {
	invLen := 1.0 / v.Len()
	if math32.IsInf(invLen, 1) {
		return Vec2Zero()
	}
	return Vector2{v.X * invLen, v.Y * invLen}
}

// Returns the negative of the vector
func (v Vector2) Neg() Vector2 {
	return Vector2{-v.X, -v.Y}
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

// Returns the element-wise product of a vector and a vector
func (v1 Vector2) MulComponent(v2 Vector2) Vector2 {
	return Vector2{v1.X * v2.X, v1.Y * v2.Y}
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

// Performs a min comparison for all the elements of the vector
func (v Vector2) ComponentMin(c float32) Vector2 {
	return Vector2{math32.Min(c, v.X), math32.Min(c, v.Y)}
}

// Performs a max comparison for all the elements of the vector
func (v Vector2) ComponentMax(c float32) Vector2 {
	return Vector2{math32.Max(c, v.X), math32.Max(c, v.Y)}
}

// Performs a clamp operation on all the components of the vector
func (v Vector2) ComponentClamp(min, max float32) Vector2 {
	return v.ComponentMax(min).ComponentMin(max)
}

// Returns a vector with -1 or +1 corresponding to the sign of each component
func (v Vector2) SignVec() Vector2 {
	return Vector2{math32.Copysign(1, v.X), math32.Copysign(1, v.Y)}
}

// Returns the max component of a 2D vector
func (v Vector2) Max() float32 {
	return max(v.X, v.Y)
}

// Returns the min component of a 2D vector
func (v Vector2) Min() float32 {
	return min(v.X, v.Y)
}

// Returns the absolute value of a 2D vector
func (v Vector2) Abs() Vector2 {
	return Vector2{math32.Abs(v.X), math32.Abs(v.Y)}
}

// Returns the inverse of a 2D vector
func (v Vector2) Inv() Vector2 {
	return Vector2{1.0 / v.X, 1.0 / v.Y}
}

// End Vector2 Functions

// Start Vector3 Functions

// Creates a 3D Vector from components
func Vec3(x, y, z float32) Vector3 {
	return Vector3{X: x, Y: y, Z: z}
}

// Creates the 3D zero vector
func Vec3Zero() Vector3 {
	return Vector3{X: 0, Y: 0, Z: 0}
}

// Creates a 3D vector with similar components
func Vec3Splat(c float32) Vector3 {
	return Vector3{X: c, Y: c, Z: c}
}

// Creates an x unit vector
func Vec3X() Vector3 {
	return Vector3{X: 1, Y: 0, Z: 0}
}

// Creates an y unit vector
func Vec3Y() Vector3 {
	return Vector3{X: 0, Y: 1, Z: 0}
}

// Creates an z unit vector
func Vec3Z() Vector3 {
	return Vector3{X: 0, Y: 0, Z: 1}
}

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

// Returns either the normalized vector or the zero vector if the length is too small
func (v Vector3) NormalizedOrZero() Vector3 {
	invLen := 1.0 / v.Len()
	if math32.IsInf(invLen, 1) {
		return Vec3Zero()
	}
	return Vector3{v.X * invLen, v.Y * invLen, v.Z * invLen}
}

// Returns the negative of the vector
func (v Vector3) Neg() Vector3 {
	return Vector3{-v.X, -v.Y, -v.Z}
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

// Returns the element-wise product of a vector and a vector
func (v1 Vector3) MulComponent(v2 Vector3) Vector3 {
	return Vector3{v1.X * v2.X, v1.Y * v2.Y, v1.Z * v2.Z}
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

// Performs a min comparison for all the elements of the vector
func (v Vector3) ComponentMin(c float32) Vector3 {
	return Vector3{math32.Min(c, v.X), math32.Min(c, v.Y), math32.Min(c, v.Z)}
}

// Performs a max comparison for all the elements of the vector
func (v Vector3) ComponentMax(c float32) Vector3 {
	return Vector3{math32.Max(c, v.X), math32.Max(c, v.Y), math32.Max(c, v.Z)}
}

// Performs a clamp operation on all the components of the vector
func (v Vector3) ComponentClamp(min, max float32) Vector3 {
	return v.ComponentMax(min).ComponentMin(max)
}

// Returns a vector with -1 or +1 corresponding to the sign of each component
func (v Vector3) SignVec() Vector3 {
	return Vector3{math32.Copysign(1, v.X), math32.Copysign(1, v.Y), math32.Copysign(1, v.Z)}
}

// Returns the max component of a 3D vector
func (v Vector3) Max() float32 {
	return max(v.X, v.Y, v.Z)
}

// Returns the min component of a 3D vector
func (v Vector3) Min() float32 {
	return min(v.X, v.Y, v.Z)
}

// Returns the absolute value of a 3D vector
func (v Vector3) Abs() Vector3 {
	return Vector3{math32.Abs(v.X), math32.Abs(v.Y), math32.Abs(v.Z)}
}

// Returns the inverse of a 3D vector
func (v Vector3) Inv() Vector3 {
	return Vector3{1.0 / v.X, 1.0 / v.Y, 1.0 / v.Z}
}

// End Vector3 Functions
