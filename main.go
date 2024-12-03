package main

import (
	"fmt"
	"log/slog"
	"os"
	"slices"
	"strings"
)

const (
	basePathEnv    = "BASE_PATH"
	baseURLEnv     = "BASE_URL"
	repoURLTypeEnv = "REPO_URL_TYPE"
	tokenEnv       = "TOKEN"
	orderByEnv     = "ORDER_BY"
	installEnv     = "INSTALL"
	dirPermission  = 0o766
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
	repoURLType := os.Getenv(repoURLTypeEnv)
	token := os.Getenv(tokenEnv)
	orderBy := os.Getenv(orderByEnv)
	install := os.Getenv(installEnv)

	if basePath == "" || baseURL == "" || token == "" {
		slog.Error("mandatory envs have no value")

		return
	}

	if strings.ToLower(repoURLType) == repoURLTypeHTTP {
		repoURLType = repoURLTypeHTTP
	}

	orderByItems := []string{"id", "name", "path", "created_at", "updated_at", "last_activity_at"}
	if orderBy == "" || !slices.Contains(orderByItems, orderBy) {
		orderBy = "id"
	}

	installPackage := false
	if strings.ToLower(install) == "true" {
		installPackage = true
	}

	projects := getProjects(baseURL, repoURLType, token, orderBy)

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

			gitOut, gitErr = pull(projectPath)
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
