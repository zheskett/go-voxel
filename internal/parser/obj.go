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
	Vertices []te.Vector3
	Edges    [][2]int
	Faces    [][3]int
	// Don't use uv/normals (yet?)
}

// ParseObj returns an Obj from object file
func ParseObj(path string) (Obj, error) {
	obj := Obj{}
	edgeSet := make(map[[2]int]bool)

	file, err := os.Open(path)
	if err != nil {
		return obj, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1
	absLargestVertPos := float32(0)
	for scanner.Scan() {
		line := scanner.Text()
		line = strings.TrimSpace(line)
		if len(line) < 2 {
			continue
		}

		switch line[:2] {
		case "v ":
			vert, err := parseVertex(line)
			absLargestVertPos = math32.Max(absLargestVertPos, math32.Max(math32.Abs(vert.X),
				math32.Max(math32.Abs(vert.Y), math32.Abs(vert.Z))))
			if err != nil {
				return obj, ObjParseError{lineNum, err}
			}
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

	obj.scale(absLargestVertPos)
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

// Scales the obj data so that the largest vert pos is at 1 (or -1)
func (obj *Obj) scale(absLargestVertPos float32) {
	scaleFactor := 1.0 / absLargestVertPos
	for i := range obj.Vertices {
		obj.Vertices[i] = obj.Vertices[i].Mul(scaleFactor)
	}
}
