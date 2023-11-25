//go:build windows

package ssh

import "testing"

func Test_BaseCMD(t *testing.T) {
	expected := "cmd /c ssh"
	result := BaseCMD()

	if result != expected {
		t.Errorf("Expected '%s' but got '%s'", expected, result)
	}
}
