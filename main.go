package main

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strconv"
	"strings"
)

const (
	basePathEnv                     = "BASE_PATH"
	baseURLEnv                      = "BASE_URL"
	baseGroupIDEnv                  = "BASE_GROUP_ID"
	repoURLTypeEnv                  = "REPO_URL_TYPE"
	tokenEnv                        = "TOKEN"
	orderByEnv                      = "ORDER_BY"
	pullAllBranchesEnv              = "PULL_ALL_BRANCHES"
	setContainerExpirationPolicyEnv = "SET_CONTAINER_EXPIRATION_POLICY"
	installEnv                      = "INSTALL"
	dirPermission                   = 0o766
)

var (
	Version   string
	BuildDate string
)

func main() {
	slog.Info("gitlab pull",
		"version", Version,
		"build_date", BuildDate,
	)

	basePath := os.Getenv(basePathEnv)
	baseURL := os.Getenv(baseURLEnv)
	baseGroupID := os.Getenv(baseGroupIDEnv)
	repoURLType := os.Getenv(repoURLTypeEnv)
	token := os.Getenv(tokenEnv)
	orderBy := os.Getenv(orderByEnv)
	pullAllBranches := os.Getenv(pullAllBranchesEnv)
	setContainerExpirationPolicy := os.Getenv(setContainerExpirationPolicyEnv)
	install := os.Getenv(installEnv)

	if basePath == "" || baseURL == "" || token == "" {
		slog.Error("mandatory envs have no value")

		return
	}

	_, err := strconv.Atoi(baseGroupID)
	if err != nil {
		baseGroupID = ""
	}

	if strings.ToLower(repoURLType) == repoURLTypeHTTP {
		repoURLType = repoURLTypeHTTP
	}

	orderByItems := []string{"id", "name", "path", "created_at", "updated_at", "last_activity_at"}
	if orderBy == "" || !slices.Contains(orderByItems, orderBy) {
		orderBy = "id"
	}

	pullAllBranchesStatus := false
	if strings.ToLower(pullAllBranches) == "true" {
		pullAllBranchesStatus = true
	}

	setContainerExpirationPolicyAttr := false
	if strings.ToLower(setContainerExpirationPolicy) == "true" {
		setContainerExpirationPolicyAttr = true
	}

	installPackage := false
	if strings.ToLower(install) == "true" {
		installPackage = true
	}

	projects := getProjects(baseURL, baseGroupID, repoURLType, token, orderBy, setContainerExpirationPolicyAttr)

	projectsCount := len(projects)
	if projectsCount == 0 {
		slog.Info("no projects found")

		return
	}

	slog.Info(fmt.Sprintf("%d projects found", projectsCount))

	// Create base directory
	dirErr := os.MkdirAll(basePath, dirPermission)
	if dirErr != nil {
		slog.Error(dirErr.Error())

		return
	}

	for i, project := range projects {
		fmt.Println()

		slog.Info("processing project",
			"index", i,
			"path", project.PathWithNamespace,
		)

		// check project dir
		projectPath := fmt.Sprintf("%s/%s", basePath, project.PathWithNamespace)
		_, projectPathErr := os.Stat(projectPath)

		var (
			gitOut []byte
			gitErr error
		)

		if projectPathErr == nil {
			slog.Info("git pull")

			gitOut, gitErr = pull(projectPath, pullAllBranchesStatus)
		} else {
			slog.Info("git clone")

			repoURL := getRepoURL(repoURLType, project)
			gitOut, gitErr = clone(repoURL, projectPath)
		}

		if gitErr != nil {
			slog.Error(gitErr.Error())
		}

		if len(gitOut) > 0 {
			slog.Info(string(gitOut))
		}

		// Get packages
		if installPackage {
			packageOut, packageErr := getPackage(projectPath)

			if packageErr != nil {
				slog.Error(packageErr.Error())
			}

			if len(packageOut) > 0 {
				slog.Info(string(packageOut))
			}
		}
	}
}
