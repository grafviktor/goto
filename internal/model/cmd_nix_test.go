//go:build !windows

package model

import "testing"

func TestBaseCMD(t *testing.T) {
	expected := "ssh"
	result := BaseCMD()

	if result != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, result)
	}
}