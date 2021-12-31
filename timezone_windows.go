//go:build windows

package timezone

import (
	"fmt"
	"runtime"
)

// Name always return an error as it's not implemented yet for current os.
func Name() (string, error) {
	return "", fmt.Errorf("name not implemented")
}
