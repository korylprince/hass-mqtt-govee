package main

import (
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
)

func NewCmd(path string, extraArgs ...string) (io.Reader, error) {
	defArgs := []string{"-R", "192", "-F", "json", "-M", "time:iso:tz:local"}
	args := append(defArgs, extraArgs...)
	cmd := exec.Command(path, args...)
	cmd.Stderr = os.Stderr
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, fmt.Errorf("could not get stdout pipe: %w", err)
	}
	if err = cmd.Start(); err != nil {
		return nil, fmt.Errorf("could not start process: %w", err)
	}
	log.Println("rtl_433: started successfully")

	go func() {
		if err := cmd.Wait(); err != nil {
			log.Println("rtl_433: exited with nonzero status:", err)
			return
		}
		log.Println("rtl_433: exited successfully")
	}()

	return stdout, nil
}
