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

func pull(dir string, pullAllBranches bool) ([]byte, error) {
	err := os.Chdir(dir)
	if err != nil {
		return []byte{}, err
	}

	// get all branches
	if pullAllBranches {
		allBranches, errAllBranches := getAllBranches()
		if errAllBranches != nil {
			return []byte{}, errAllBranches
		}

		for _, branch := range allBranches {
			errSwitchBranch := switchBranch(branch)
			if errSwitchBranch != nil {
				continue
			}

			// git pull
			cmdPull := exec.Command("git", "pull")

			errPull := cmdPull.Run()
			if errPull != nil {
				continue
			}
		}
	}

	// get default branch
	defaultBranch, errorDefaultBranch := getDefaultBranch()
	if errorDefaultBranch != nil {
		return []byte{}, errorDefaultBranch
	}

	errSwitchBranch := switchBranch(defaultBranch)
	if errSwitchBranch != nil {
		return []byte{}, errSwitchBranch
	}

	// git fetch all
	cmdFetch := exec.Command("git", "fetch", "--all")

	errFetch := cmdFetch.Run()
	if errFetch != nil {
		return []byte{}, errFetch
	}

	// git pull all
	cmd := exec.Command("git", "pull", "--all")

	return cmd.Output()
}

func switchBranch(branch string) error {
	cmd := exec.Command("git", "switch", branch)

	_, err := cmd.Output()

	return err
}

func getDefaultBranch() (string, error) {
	cmd := exec.Command("git", "symbolic-ref", "refs/remotes/origin/HEAD", "--short")

	out, err := cmd.Output()
	if err != nil {
		return "", err
	}

	outStr := string(out)
	outStr = strings.TrimSpace(outStr)
	branch := strings.TrimLeft(outStr, "origin/")

	return branch, nil
}

func getAllBranches() ([]string, error) {
	cmd := exec.Command("git", "branch", "-r")

	out, err := cmd.Output()
	if err != nil {
		return nil, err
	}

	outStr := string(out)
	outStr = strings.TrimSpace(outStr)

	branches := []string{}

	for _, branch := range strings.Split(outStr, "\n") {
		branchName, after, found := strings.Cut(branch, "->")
		if found {
			branchName = after
		}

		branchName = strings.TrimSpace(branchName)
		if branchName == "" {
			continue
		}

		branchName = strings.TrimLeft(branchName, "origin/")

		branches = append(branches, branchName)
	}

	return branches, nil
}
