package fs

import (
	"runtime"
	"os/exec"
	"fmt"
)

func Launch(name string) error {
	switch runtime.GOOS {
	case "linux":
		return exec.Command("xdg-open", name).Start()
	case "windows", "darwin":
		return exec.Command("open", name).Start()
	default:
		return fmt.Errorf("unsupported platform: %v", runtime.GOOS)
	}
}
