//go:build !windows

package ssh

import "testing"

func TestBaseCMD(t *testing.T) {
	expected := "ssh"
	result := BaseCMD()

	if result != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, result)
	}
}
