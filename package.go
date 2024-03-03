package main

import (
	"os"
	"os/exec"
)

func getPackage(dir string) ([]byte, error) {
	var cmd *exec.Cmd

	err := os.Chdir(dir)
	if err != nil {
		return nil, err
	}

	_, err = os.Stat("go.mod")
	if err == nil {
		cmd = exec.Command("go", "mod", "tidy")
	}

	_, err = os.Stat("composer.json")
	if err == nil {
		cmd = exec.Command("composer", "install")
	}

	if cmd != nil {
		return cmd.Output()
	}

	return nil, nil
}
