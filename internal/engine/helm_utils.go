package engine

import (
	"os/exec"
	"strings"
	"sync"
)

var (
	helmDiffOnce      sync.Once
	helmDiffAvailable bool
)

// HelmDiffAvailable reports whether the helm diff plugin appears to be installed
func HelmDiffAvailable() bool {
	helmDiffOnce.Do(func() {
		if _, err := exec.LookPath("helm"); err != nil {
			return
		}

		if output, err := exec.Command("helm", "plugin", "list").CombinedOutput(); err == nil {
			if strings.Contains(string(output), "diff") {
				helmDiffAvailable = true
				return
			}
		}

		if err := exec.Command("helm", "diff", "version").Run(); err == nil {
			helmDiffAvailable = true
		}
	})
	return helmDiffAvailable
}
