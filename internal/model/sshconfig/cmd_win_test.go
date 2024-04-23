//go:build windows

package sshconfig

import "testing"

func TestBaseCMD(t *testing.T) {
	expected := "cmd /c ssh"
	result := BaseCMD()

	if result != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, result)
	}
}
