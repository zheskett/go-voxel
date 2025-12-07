package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

	"github.com/chewxy/math32"
	te "github.com/zheskett/go-voxel/internal/tensor"
)

type ObjParseError struct {
	lineNum  int
	errorMsg error
}

func (e ObjParseError) Error() string {
	return fmt.Sprintf("Error Parsing Obj File (Line %v): %v", e.lineNum, e.errorMsg)
}

// Contains information about an object
type Obj struct {
	Vertices    []te.Vector3
	Edges       [][2]int
	Faces       [][3]int
	MaxVertsPos te.Vector3
	// Don't use uv/normals (yet?)
}

// ParseObj returns an Obj from object file.
// flipX, flipY, and flipZ flip the object on the respective axis.
func ParseObj(path string, flipX, flipY, flipZ bool) (Obj, error) {
	obj := Obj{}
	edgeSet := make(map[[2]int]bool)

	file, err := os.Open(path)
	if err != nil {
		return obj, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1
	maxVertsPos := te.Vec3Splat(math32.Inf(-1))
	minVertsPos := te.Vec3Splat(math32.Inf(1))
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) < 2 {
			continue
		}

		switch line[:2] {
		case "v ":
			vert, err := parseVertex(line)
			if err != nil {
				return obj, ObjParseError{lineNum, err}
			}
			maxVertsPos = te.Vec3(max(maxVertsPos.X, vert.X), max(maxVertsPos.Y, vert.Y), max(maxVertsPos.Z, vert.Z))
			minVertsPos = te.Vec3(min(minVertsPos.X, vert.X), min(minVertsPos.Y, vert.Y), min(minVertsPos.Z, vert.Z))

			obj.Vertices = append(obj.Vertices, vert)
		case "f ":
			faces, err := parseFace(line)
			if err != nil {
				return obj, ObjParseError{lineNum, err}
			}
			for i := range faces {
				for j := range faces[i] {
					if faces[i][j] < 0 {
						faces[i][j] = len(obj.Vertices) + faces[i][j] + 1
					}
				}
				// Don't duplicate edges
				for j := range len(faces[i]) - 1 {
					for k := j + 1; k < len(faces[i]); k++ {
						v1 := min(faces[i][j], faces[i][k])
						v2 := max(faces[i][j], faces[i][k])
						if !edgeSet[[2]int{v1, v2}] {
							edgeSet[[2]int{v1, v2}] = true
							obj.Edges = append(obj.Edges, [2]int{v1, v2})
						}
					}
				}
			}
			obj.Faces = append(obj.Faces, faces...)
		default:
			continue
		}
	}

	obj.scale(maxVertsPos, minVertsPos, flipX, flipY, flipZ)
	return obj, nil
}

// Returns a Vector3 representing a vertex
func parseVertex(line string) (te.Vector3, error) {
	vertex := te.Vec3(0, 0, 0)
	parts := strings.Split(line, " ")
	// One part contains "v"
	if len(parts) < 4 {
		return vertex, errors.New("Too few vertex positions")
	}

	// Ignore "w"
	x, err := strconv.ParseFloat(parts[1], 32)
	if err != nil {
		return vertex, errors.New("Failed to parse vertex x pos")
	}
	y, err := strconv.ParseFloat(parts[2], 32)
	if err != nil {
		return vertex, errors.New("Failed to parse vertex y pos")
	}
	z, err := strconv.ParseFloat(parts[3], 32)
	if err != nil {
		return vertex, errors.New("Failed to parse vertex z pos")
	}
	vertex = te.Vec3(float32(x), float32(y), float32(z))

	return vertex, nil
}

// Returns a list of indices of vertices where faces occur
func parseFace(line string) ([][3]int, error) {
	var faces [][3]int = nil

	parts := strings.Split(line, " ")
	// One part contains "f"
	if len(parts) < 4 {
		return faces, errors.New("Too few face indices")
	}

	// Remove normal/texture info
	for i := 1; i < len(parts); i++ {
		parts[i] = strings.Split(parts[i], "/")[0]
	}

	first, err := strconv.Atoi(parts[1])
	if err != nil {
		return faces, errors.New("Failed to parse face")
	}
	second, err := strconv.Atoi(parts[2])
	if err != nil {
		return faces, errors.New("Failed to parse face")
	}

	for i := 3; i < len(parts); i++ {
		vertIdx, err := strconv.Atoi(parts[i])
		if err != nil {
			return faces, errors.New("Failed to parse face")
		}
		// 1-indexed in obj file
		faces = append(faces, [3]int{first - 1, second - 1, vertIdx - 1})
	}

	return faces, nil
}

// Positions the obj data so that the origin is at the center of the object.
// Scales the obj data so that the largest vert pos is at 1 (or -1).
// flipX, flipY, and flipZ flip the object on the respective axis.
func (obj *Obj) scale(maxVertsPos, minVertsPos te.Vector3, flipX, flipY, flipZ bool) {
	flipVec := te.Vec3(1, 1, 1)
	if flipX {
		flipVec.X = -1
	}
	if flipY {
		flipVec.Y = -1
	}
	if flipZ {
		flipVec.Z = -1
	}

	offsetVec := maxVertsPos.Add(minVertsPos).Mul(0.5)
	maxAbsPos := maxVertsPos.Sub(offsetVec).Max()

	scaleFactor := 1.0 / maxAbsPos
	obj.MaxVertsPos = maxVertsPos.Sub(offsetVec).Mul(scaleFactor).ComponentMin(1.0)

	// Translate each point by the offset and scale
	for i, v := range obj.Vertices {
		obj.Vertices[i] = v.Sub(offsetVec).Mul(scaleFactor).MulComponent(flipVec).ComponentClamp(-1.0, 1.0)
	}
}
