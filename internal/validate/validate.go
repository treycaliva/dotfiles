package validate

import (
	"bytes"
	"os/exec"
)

// Result holds the outcome of a post-install validation command.
type Result struct {
	Package string
	Skipped bool
	Output  string
	Err     error
}

// Run executes a validation command for a package.
// If the command string is empty, returns a skipped result.
func Run(pkg, command string) Result {
	if command == "" {
		return Result{Package: pkg, Skipped: true}
	}

	cmd := exec.Command("sh", "-c", command)
	var buf bytes.Buffer
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	err := cmd.Run()

	return Result{
		Package: pkg,
		Output:  buf.String(),
		Err:     err,
	}
}
