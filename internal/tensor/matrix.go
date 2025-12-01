package tensor

import "github.com/chewxy/math32"

// A Matrix is stored in column-major order
type Matrix interface {
	// Returns the number of rows in this matrix
	Rows()
	// Returns the number of columns in this matrix
	Cols()
	// Returns the element at the given row and column
	At(row, col int) float32
	// Returns the determinant of a matrix
	Det() float32
}

type Matrix2x2 [2 * 2]float32
type Matrix3x3 [3 * 3]float32

// Start Matrix2x2 Functions
// Start Matrix2x2 implements Matrix

// Returns the number of rows in this matrix (2)
func (m Matrix2x2) Rows() int {
	return 2
}

// Returns the number of columns in this matrix (2)
func (m Matrix2x2) Cols() int {
	return 2
}

// Returns the element at the given row and column.
// This is 0-indexed.
// Equivalent to mat[col * 2 + row]
func (m Matrix2x2) At(row, col int) float32 {
	return m[col*2+row]
}

// Returns the determinant of a matrix
func (m Matrix2x2) Det() float32 {
	return m[0]*m[3] - m[1]*m[2]
}

// End Matrix2x2 implements Matrix

// Returns the row at the given index
func (m Matrix2x2) Row(row int) Vector2 {
	return Vector2{m[row], m[2+row]}
}

// Returns the column at the given index
func (m Matrix2x2) Col(col int) Vector2 {
	return Vector2{m[col*2], m[col*2+1]}
}

// Matrix2x2FromRows builds a new matrix from row vectors.
func Matrix2x2FromRows(v1, v2 Vector2) Matrix2x2 {
	return Matrix2x2{v1.X, v2.X, v1.Y, v2.Y}
}

// Matrix2x2FromCols builds a new matrix from column vectors.
func Matrix2x2FromCols(v1, v2 Vector2) Matrix2x2 {
	return Matrix2x2{v1.X, v1.Y, v2.X, v2.Y}
}

// Rotate2D returns a 2D rotation matrix
func Rotate2D(angle float32) Matrix2x2 {
	sin, cos := math32.Sincos(angle)
	return Matrix2x2{cos, sin, -sin, cos}
}

func (m1 Matrix2x2) Add(m2 Matrix2x2) Matrix2x2 {
	return Matrix2x2{
		m1[0] + m2[0],
		m1[1] + m2[1],
		m1[2] + m2[2],
		m1[3] + m2[3],
	}
}

func (m1 Matrix2x2) Mul(m2 Matrix2x2) Matrix2x2 {
	return Matrix2x2{
		m1[0]*m2[0] + m1[2]*m2[1],
		m1[1]*m2[0] + m1[3]*m2[1],
		m1[0]*m2[2] + m1[2]*m2[3],
		m1[1]*m2[2] + m1[3]*m2[3],
	}
}

func (m Matrix2x2) MulVec(v Vector2) Vector2 {
	return Vector2{
		m[0]*v.X + m[2]*v.Y,
		m[1]*v.X + m[3]*v.Y,
	}
}

// End Matrix2x2 Functions

// Start Matrix3x3 Functions
// Start Matrix3x3 implements Matrix

// Returns the number of rows in this matrix (3)
func (m Matrix3x3) Rows() int {
	return 3
}

// Returns the number of columns in this matrix (3)
func (m Matrix3x3) Cols() int {
	return 3
}

// Returns the element at the given row and column.
// This is 0-indexed.
// Equivalent to mat[col * 3 + row]
func (m Matrix3x3) At(row, col int) float32 {
	return m[col*3+row]
}

// Returns the determinant of a matrix
func (m Matrix3x3) Det() float32 {
	return (m[0]*m[4]*m[8] + m[3]*m[7]*m[2] + m[6]*m[1]*m[5] -
		m[6]*m[4]*m[2] - m[3]*m[1]*m[8] - m[0]*m[7]*m[5])
}

// End Matrix3x3 implements Matrix

// Returns the row at the given index
func (m Matrix3x3) Row(row int) Vector3 {
	return Vector3{m[row], m[3+row], m[6+row]}
}

// Returns the column at the given index
func (m Matrix3x3) Col(col int) Vector3 {
	return Vector3{m[col*3], m[col*3+1], m[col*3+2]}
}

// Matrix3x3FromRows builds a new matrix from row vectors.
func Matrix3x3FromRows(v1, v2, v3 Vector3) Matrix3x3 {
	return Matrix3x3{v1.X, v2.X, v3.X, v1.Y, v2.Y, v3.Y, v1.Z, v2.Z, v3.Z}
}

// Matrix3x3FromCols builds a new matrix from column vectors.
func Matrix3x3FromCols(v1, v2, v3 Vector3) Matrix3x3 {
	return Matrix3x3{v1.X, v1.Y, v1.Z, v2.X, v2.Y, v2.Z, v3.X, v3.Y, v3.Z}
}

// Rotate3DX returns a 3D rotation matrix about the X axis
//
// [1 0 0]
// [0 cos -sin]
// [0 sin cos]
func Rotate3DX(angle float32) Matrix3x3 {
	sin, cos := math32.Sincos(angle)
	return Matrix3x3{1, 0, 0, 0, cos, sin, 0, -sin, cos}
}

// Rotate3DY returns a 3D rotation matrix about the Y axis
//
// [cos 0 sin]
// [0 1 0]
// [-sin 0 cos]
func Rotate3DY(angle float32) Matrix3x3 {
	sin, cos := math32.Sincos(angle)
	return Matrix3x3{cos, 0, -sin, 0, 1, 0, sin, 0, cos}
}

// Rotate3DZ returns a 3D rotation matrix about the Z axis
//
// [cos -sin 0]
// [sin cos 0]
// [0 0 1]
func Rotate3DZ(angle float32) Matrix3x3 {
	sin, cos := math32.Sincos(angle)
	return Matrix3x3{cos, sin, 0, -sin, cos, 0, 0, 0, 1}
}

// Rotate3DXYZ returns a 3D body-sequence rotation matrix about the X -> Y -> Z axes.
// This is an intrinsic rotation.
//
// [cos(y)cos(z), -cos(y)sin(z), sin(y)]
// [cos(x)sin(z)+sin(x)sin(y)cos(z), cos(x)cos(z)-sin(x)sin(y)sin(z), -sin(x)cos(y)]
// [sin(x)sin(z)-cos(x)sin(y)cos(z), sin(x)cos(z)+cos(x)sin(y)sin(z), cos(x)cos(y)]
func Rotate3DXYZ(xAngle, yAngle, zAngle float32) Matrix3x3 {
	sinx, cosx := math32.Sincos(xAngle)
	siny, cosy := math32.Sincos(yAngle)
	sinz, cosz := math32.Sincos(zAngle)

	return Matrix3x3FromRows(
		Vector3{cosy * cosz, -cosy * sinz, siny},
		Vector3{cosx*sinz + sinx*siny*cosz, cosx*cosz - sinx*siny*sinz, -sinx * cosy},
		Vector3{sinx*sinz - cosx*siny*cosz, sinx*cosz + cosx*siny*sinz, cosx * cosy},
	)
}

// Rotate3DZYX returns a 3D body-sequence rotation matrix about the Z -> Y -> X axes.
// This is an intrinsic rotation.
//
// [cos(y)cos(z), -cos(y)sin(z), sin(y)]
// [sin(x)sin(y)cos(z)+cos(x)sin(z), -sin(x)sin(y)sin(x)+cos(x)cos(z), -sin(x)cos(y)]
// [-cos(x)sin(y)cos(z)+sin(x)sin(z), cos(x)sin(x)sin(x)+sin(y)cos(z), cos(x)cos(y)]
func Rotate3DZYX(xAngle, yAngle, zAngle float32) Matrix3x3 {
	sinx, cosx := math32.Sincos(xAngle)
	siny, cosy := math32.Sincos(yAngle)
	sinz, cosz := math32.Sincos(zAngle)

	return Matrix3x3FromRows(
		Vector3{cosy * cosz, -cosy * sinz, siny},
		Vector3{sinx*siny*cosz + cosx*sinz, -sinx*siny*sinz + cosx*cosz, -sinx * cosy},
		Vector3{-cosx*siny*cosz + sinx*sinz, cosx*siny*sinz + sinx*cosz, cosx * cosy},
	)
}

func (m1 Matrix3x3) Add(m2 Matrix3x3) Matrix3x3 {
	return Matrix3x3{
		m1[0] + m2[0],
		m1[1] + m2[1],
		m1[2] + m2[2],
		m1[3] + m2[3],
		m1[4] + m2[4],
		m1[5] + m2[5],
		m1[6] + m2[6],
		m1[7] + m2[7],
		m1[8] + m2[8],
	}
}

func (m1 Matrix3x3) Mul(m2 Matrix3x3) Matrix3x3 {
	return Matrix3x3{
		m1[0]*m2[0] + m1[3]*m2[1] + m1[6]*m2[2],
		m1[1]*m2[0] + m1[4]*m2[1] + m1[7]*m2[2],
		m1[2]*m2[0] + m1[5]*m2[1] + m1[8]*m2[2],
		m1[0]*m2[3] + m1[3]*m2[4] + m1[6]*m2[5],
		m1[1]*m2[3] + m1[4]*m2[4] + m1[7]*m2[5],
		m1[2]*m2[3] + m1[5]*m2[4] + m1[8]*m2[5],
		m1[0]*m2[6] + m1[3]*m2[7] + m1[6]*m2[8],
		m1[1]*m2[6] + m1[4]*m2[7] + m1[7]*m2[8],
		m1[2]*m2[6] + m1[5]*m2[7] + m1[8]*m2[8],
	}
}

func (m Matrix3x3) MulVec(v Vector3) Vector3 {
	return Vector3{
		m[0]*v.X + m[3]*v.Y + m[6]*v.Z,
		m[1]*v.X + m[4]*v.Y + m[7]*v.Z,
		m[2]*v.X + m[5]*v.Y + m[8]*v.Z,
	}
}

// End Matrix3x3 Functions
