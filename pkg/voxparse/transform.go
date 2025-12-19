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
	Frames      []VoxFrame
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
	transform := VoxTransform{}
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
		rot, ok := dict["_r"]
		if ok && len(rot) > 0 {
			byteRot := rot[0]
			transform.Frames[i].Rotation = transRot(byteRot)
		}

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

	return &transform, nil
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
