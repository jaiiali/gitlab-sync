package main

import (
	"os"
	"os/exec"
	"strings"
)

func clone(url, dir string) ([]byte, error) {
	cmd := exec.Command("git", "clone", url, dir, "--recursive")

	return cmd.Output()
}

func pull(dir string) ([]byte, error) {
	err := os.Chdir(dir)
	if err != nil {
		return []byte{}, err
	}

	// git branch track
	cmdBranch := exec.Command("git", "branch", "-r")

	outBranch, errBranch := cmdBranch.Output()
	if errBranch != nil {
		return []byte{}, errBranch
	}

	outBranchStr := string(outBranch)
	outBranchStr = strings.TrimSpace(outBranchStr)

	for _, branch := range strings.Split(outBranchStr, "\n") {
		branch = strings.TrimSpace(branch)
		if branch == "" {
			continue
		}

		if strings.Contains(branch, "->") {
			continue
		}

		localBranch := strings.TrimLeft(branch, "origin/")
		cmdTrack := exec.Command("git", "branch", "--track", localBranch, branch)

		errTrack := cmdTrack.Run()
		if errTrack != nil {
			continue
		}
	}

	// git fetch
	cmdFetch := exec.Command("git", "fetch", "--all")

	errFetch := cmdFetch.Run()
	if errFetch != nil {
		return []byte{}, errFetch
	}

	// git pull
	cmd := exec.Command("git", "pull", "--all")

	return cmd.Output()
}
