package voxparse

import (
	"fmt"
	"strconv"
	"strings"
)

var (
	defaultRot = VoxRotation{{1, 0, 0}, {0, 1, 0}, {0, 0, 1}}
)

type VoxTransform struct {
	NodeId      int32
	ChildNodeId int32
	LayerId     int32
	Transforms  []*VoxTransform
	Shape       *VoxShape
	Frames      []VoxFrame
}

type voxGroup struct {
	nodeId          int32
	childrenNodeIds []int32
}

type VoxShape struct {
	NodeId   int32
	ModelIds []int32
}

type VoxFrame struct {
	Rotation    VoxRotation
	Translation [3]int32
	Index       int32
}

// Row-major int rotation. Example:
//
//	0  1  0
//	0  0  -1
//	-1  0  0
type VoxRotation [3][3]int8

type voxDict map[string]string

func (fb *fileBytes) parseSceneTree() (*VoxTransform, error) {
	transforms := make(map[int32]*VoxTransform)
	groups := make(map[int32]*voxGroup)
	shapes := make(map[int32]*VoxShape)
	root, err := fb.parseTransform()
	if err != nil {
		return nil, err
	}
	transforms[root.NodeId] = root
	curr_tag := string(fb.byteArr[fb.pos : fb.pos+4])
	for len(fb.byteArr[fb.pos:]) >= 4 && (curr_tag == "nTRN" || curr_tag == "nGRP" || curr_tag == "nSHP") {
		switch curr_tag {
		case "nTRN":
			t, err := fb.parseTransform()
			if err != nil {
				return nil, err
			}
			transforms[t.NodeId] = t

		case "nGRP":
			g, err := fb.parseGroup()
			if err != nil {
				return nil, err
			}
			groups[g.nodeId] = g

		case "nSHP":
			s, err := fb.parseShape()
			if err != nil {
				return nil, err
			}
			shapes[s.NodeId] = s
		}
		curr_tag = string(fb.byteArr[fb.pos : fb.pos+4])
	}

	for _, t := range transforms {
		if g, ok := groups[t.ChildNodeId]; ok && g != nil {
			for _, gt := range g.childrenNodeIds {
				t.Transforms = append(t.Transforms, transforms[gt])
			}
		} else if s, ok := shapes[t.ChildNodeId]; ok && s != nil {
			t.Shape = s
		}
	}

	return root, nil
}

// Parse Transform Information from Scene Graph
//
// This solution is largely adapted from the dot_vox cargo crate
// https://github.com/dust-engine/dot_vox
func (fb *fileBytes) parseTransform() (*VoxTransform, error) {
	// absorb nTRN tag
	if string(fb.byteArr[fb.pos:fb.pos+4]) != "nTRN" {
		return nil, fmt.Errorf("No nTRN tag found")
	}
	fb.pos += 4
	// Seems to be 2 ints here, but I don't know what they are for
	fb.pos += 8
	transform := new(VoxTransform)
	transform.NodeId = fb.readInt()
	// ignore for now
	fb.readDict()
	transform.ChildNodeId = fb.readInt()
	reserved := fb.readInt()
	if reserved != -1 {
		return nil, fmt.Errorf("Transform reserved id must be -1, got %v", reserved)
	}
	transform.LayerId = fb.readInt()

	numFrames := fb.readInt()
	transform.Frames = make([]VoxFrame, numFrames)
	for i := range numFrames {
		dict := fb.readDict()
		transform.Frames[i].Rotation = defaultRot
		// Ignore rotation for now
		// rot, ok := dict["_r"]
		// if ok && len(rot) > 0 {
		// 	byteRot := rot[0]
		// 	transform.Frames[i].Rotation = transRot(byteRot)
		// }

		trans, ok := dict["_t"]
		if ok {
			// Tokenize trans into 3 ints
			tokens := strings.Fields(trans)
			if len(tokens) != 3 {
				return nil, fmt.Errorf("Malformed transform")
			}

			for j := range tokens {
				t, err := strconv.ParseInt(tokens[j], 10, 32)
				if err != nil {
					return nil, err
				}
				transform.Frames[i].Translation[j] = int32(t)
			}
		}

		transform.Frames[i].Index = i
	}

	return transform, nil
}

func (fb *fileBytes) parseGroup() (*voxGroup, error) {
	// absorb nGRP tag
	if string(fb.byteArr[fb.pos:fb.pos+4]) != "nGRP" {
		return nil, fmt.Errorf("No nGRP tag found")
	}
	fb.pos += 4
	group := new(voxGroup)
	// Seems to be 2 ints here, but I don't know what they are for
	fb.pos += 8
	group.nodeId = fb.readInt()
	// ignore for now
	fb.readDict()
	numChildren := fb.readInt()
	for range numChildren {
		group.childrenNodeIds = append(group.childrenNodeIds, fb.readInt())
	}

	return group, nil
}

func (fb *fileBytes) parseShape() (*VoxShape, error) {
	// absorb nSHP tag
	if string(fb.byteArr[fb.pos:fb.pos+4]) != "nSHP" {
		return nil, fmt.Errorf("No nSHP tag found")
	}
	fb.pos += 4
	shape := new(VoxShape)
	// Seems to be 2 ints here, but I don't know what they are for
	fb.pos += 8
	shape.NodeId = fb.readInt()
	// ignore for now
	fb.readDict()
	numModels := fb.readInt()

	for range numModels {
		shape.ModelIds = append(shape.ModelIds, fb.readInt())
		// ignore dict
		fb.readDict()
	}

	return shape, nil
}

// readString reads the .vox string type, seeks to the next position in the file, the reutns the string
// Does not do bounds checking!
func (fb *fileBytes) readString() string {
	bufSize := fb.readInt()
	// No null terminator in provided strings
	byteStr := fb.byteArr[fb.pos : fb.pos+int(bufSize)]
	fb.pos += int(bufSize)

	return string(byteStr)
}

// Read a dictionary entry and seek to next position
// Does not do bounds checking!
func (fb *fileBytes) readDict() voxDict {
	numPairs := fb.readInt()

	dict := make(voxDict, numPairs)
	for range numPairs {
		key := fb.readString()
		val := fb.readString()
		dict[key] = val
	}
	return dict
}

// Read a rotation and seek to next position in the file
// Does not do bounds checking!
func transRot(val byte) VoxRotation {
	var rowIdxs [3]byte
	rowIdxs[0] = val & 0x03 // Last 2 bits
	rowIdxs[1] = val & 0x0c // Bits 2-3
	rowIdxs[2] = 3 - rowIdxs[0] - rowIdxs[1]
	rot := VoxRotation{}
	rot[0][rowIdxs[0]] = -int8(val & 0x10) // Bit 4
	rot[1][rowIdxs[1]] = -int8(val & 0x20) // Bit 5
	rot[2][rowIdxs[2]] = -int8(val & 0x40) // Bit 6

	return rot
}
