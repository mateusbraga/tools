package executil

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

var verboseError = `
err cmd '%+v': %v
stdout: --------------
%v
stderr: --------------
%v
`

func RunWithVerboseError(cmd *exec.Cmd) (string, error) {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		verr := fmt.Errorf(verboseError, cmd, err, stdout, stderr)
		return "", verr
	}
	return stdout.String(), nil
}

func MustRun(cmd *exec.Cmd) string {
	output, err := RunWithVerboseError(cmd)
	if err != nil {
		log.Panicln(err)
	}
	return output
}
