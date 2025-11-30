package parser

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strconv"
	"strings"

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
	Vertices     []te.Vector3
	FaceVertices [][3]int
	// Don't use uv/normals (yet?)
}

func ParseObj(path string) (Obj, error) {
	obj := Obj{}

	file, err := os.Open(path)
	if err != nil {
		return obj, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 1
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
			obj.Vertices = append(obj.Vertices, vert)
		case "f ":
			faces, err := parseFace(line)
			if err != nil {
				return obj, ObjParseError{lineNum, err}
			}
			obj.FaceVertices = append(obj.FaceVertices, faces...)
		default:
			continue
		}
	}

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
