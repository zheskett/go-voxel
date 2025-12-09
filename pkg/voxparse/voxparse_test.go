// SPDX-License-Identifier: CC0-1.0

package voxparse

import (
	"testing"
)

// TestRejectBadFileHeader calls voxparse.Parse with a bad file header and
// expects it to return an error.
func TestRejectBadFileHeader(t *testing.T) {
	_, err := Parse("assets/bunny.obj")
	if err == nil {
		t.Errorf("Expected error, got nil")
	}
}
