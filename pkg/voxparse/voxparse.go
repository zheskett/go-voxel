// SPDX-License-Identifier: CC0-1.0

// Package voxparse provides the ability to parse .vox files with the Parse() method.
// Information about the vox file format can be found here: https://paulbourke.net/dataformats/vox/
package voxparse

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"os"
)

const (
	voxMagicString = "VOX "
	mainTag        = "MAIN"
	packTag        = "PACK"
	sizeTag        = "SIZE"
	xyziTag        = "XYZI"
	// colorTag       = "RGBA" (Use default colors for now)
)

type fileBytes struct {
	byteArr []byte
	pos     int
}

// Vox contains information about a .vox file.
type Vox struct {
	Version   int     // The version of the .vox file
	NumModels int     // The number of models
	Models    []Model // The model data
}

// Models contains the size of a model and the model data
type Model struct {
	SizeX, SizeY, SizeZ int
	Voxels              []XYZI
}

// XYZI contains the X, Y, Z, and color index values of a voxel
type XYZI struct {
	X, Y, Z, I byte // X, Y, Z, ColorIndex
}

// Parse parses a .vox file and returns a Vox object.
func Parse(path string) (Vox, error) {
	vox := Vox{}
	byteArr, err := os.ReadFile(path)
	if err != nil {
		return Vox{}, err
	}
	fb := fileBytes{byteArr, 0}

	vox.Version, err = fb.parseHeader()
	if err != nil {
		return Vox{}, err
	}

	vox.NumModels, err = fb.checkPack()
	if err != nil {
		return Vox{}, err
	}

	for range vox.NumModels {
		model, err := fb.parseModel()
		if err != nil {
			return Vox{}, err
		}
		vox.Models = append(vox.Models, model)
	}

	fmt.Printf("%s\n", fb.byteArr[fb.pos:])

	return vox, nil
}

// readInt reads an integer, seeks to the next position in the file, then returns the integer
// Does not do bounds checking!
func (fb *fileBytes) readInt() int {
	val := int(binary.LittleEndian.Uint32(fb.byteArr[fb.pos:]))
	fb.pos += 4
	return val
}

// findTag seeks to the right after the next occurrence of a tag and its metadata.
// Returns the chunk data size and children chunks
// Returns an error if the tag is not found
func (fb *fileBytes) findTag(tag string) (int, int, error) {
	location := bytes.Index(fb.byteArr, []byte(tag))
	if location == -1 {
		return 0, 0, fmt.Errorf("Tag %v not found in file", tag)
	}

	fb.pos = location + len(tag)
	if fb.pos >= len(fb.byteArr)-8 {
		return 0, 0, fmt.Errorf("Tag %v data occurs passed file end", tag)
	}
	chunkSize := fb.readInt()
	childrenSize := fb.readInt()
	return chunkSize, childrenSize, nil
}

// parseHeader parses the header of a .vox file.
// Also parses the MAIN tag to check validity
// Returns the version and returns error if one has occurred
func (fb *fileBytes) parseHeader() (int, error) {
	if string(fb.byteArr[:len(voxMagicString)]) != voxMagicString {
		return 0, fmt.Errorf("Invalid vox file, magic string not found")
	}
	fb.pos = len(voxMagicString) + 4
	version := int(binary.LittleEndian.Uint32(fb.byteArr[len(voxMagicString):]))

	mainSize, mainChildrenSize, err := fb.findTag(mainTag)
	if err != nil {
		return 0, err
	}
	if mainSize != 0 || mainChildrenSize != len(fb.byteArr)-fb.pos {
		return 0, fmt.Errorf("Malformed main tag")
	}

	return version, nil
}

// checkPack returns the number of models via the PACK header
func (fb *fileBytes) checkPack() (int, error) {
	if len(fb.byteArr[fb.pos:]) < 16 || string(fb.byteArr[fb.pos:fb.pos+4]) != packTag {
		return 1, nil
	}

	_, _, err := fb.findTag(packTag)
	if err != nil {
		return 0, err
	}

	models := fb.readInt()
	return models, nil
}

// parseModel parses the model data of a .vox file.
// Returns a model and an error if one has occurred
func (fb *fileBytes) parseModel() (Model, error) {
	if len(fb.byteArr[fb.pos:]) < 40 ||
		string(fb.byteArr[fb.pos:fb.pos+4]) != sizeTag ||
		string(fb.byteArr[fb.pos+24:fb.pos+28]) != xyziTag {
		return Model{}, fmt.Errorf("Malformed model data metadata")
	}

	_, _, err := fb.findTag(sizeTag)
	if err != nil {
		return Model{}, err
	}

	model := Model{}
	model.SizeX = fb.readInt()
	model.SizeY = fb.readInt()
	model.SizeZ = fb.readInt()

	_, _, err = fb.findTag(xyziTag)
	if err != nil {
		return Model{}, err
	}
	numVoxels := fb.readInt()
	model.Voxels = make([]XYZI, numVoxels)

	for i := range numVoxels {
		model.Voxels[i].X = fb.byteArr[fb.pos]
		model.Voxels[i].Y = fb.byteArr[fb.pos+1]
		model.Voxels[i].Z = fb.byteArr[fb.pos+2]
		model.Voxels[i].I = fb.byteArr[fb.pos+3]
		fb.pos += 4
	}

	return model, nil
}
