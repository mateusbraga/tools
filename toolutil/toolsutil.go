package toolutil

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
)

func VerboseCmdRun(cmd *exec.Cmd) error {
	var stdout bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	if err := cmd.Run(); err != nil {
		log.Println("cmd:", cmd)
		log.Printf("stdout:\n---Start---\n%v\n---End---", stdout.String())
		log.Printf("stderr:\n---Start---\n%v\n---End---", stderr.String())
		return fmt.Errorf("cmd '%v': %v", cmd, err)
	}
	return nil
}
