package main

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"time"
)

const (
	maxErrCount           = 20
	pageSize              = 20
	privateTokenHeader    = "PRIVATE-TOKEN"
	contentTypeHeader     = "Content-Type"
	applicationJSONHeader = "application/json"
)

func getProjects(baseURL, baseGroupID, repoURLType, token, orderBy string, setContainerExpirationPolicyAttr bool) []Projects {
	var (
		pageNumber int
		projects   []Projects
	)

	errCount := 0

	for {
		if errCount >= maxErrCount {
			break
		}

		pageNumber++
		url := fmt.Sprintf("%s/api/v4/projects?order_by=%s&sort=desc&per_page=%d&page=%d",
			baseURL,
			orderBy,
			pageSize,
			pageNumber,
		)

		if baseGroupID != "" {
			url = fmt.Sprintf("%s/api/v4/groups/%s/projects?order_by=%s&sort=desc&per_page=%d&page=%d",
				baseURL,
				baseGroupID,
				orderBy,
				pageSize,
				pageNumber,
			)
		}

		respBody, err := do(http.MethodGet, url, token, nil)
		if err != nil {
			slog.Error("http error",
				"error", err.Error(),
				"url", url,
			)

			errCount++

			continue
		}

		var currentProjects []Projects

		err = json.Unmarshal(respBody, &currentProjects)
		if err != nil {
			slog.Error("unmarshal error",
				"error", err.Error(),
				"url", url,
			)

			errCount++

			continue
		}

		if len(currentProjects) > 0 {
			projects = append(projects, currentProjects...)

			if setContainerExpirationPolicyAttr {
				for _, project := range currentProjects {
					hasReg, hasRegistryErr := hasRegistry(baseURL, token, project)
					if hasRegistryErr != nil {
						slog.Error(hasRegistryErr.Error())
					}

					if hasReg {
						setCleanupErr := setCleanup(baseURL, token, project)
						if setCleanupErr != nil {
							slog.Error(setCleanupErr.Error())
						}
					}
				}
			}

			for _, project := range currentProjects {
				wikiProject, wikiErr := getWikis(baseURL, repoURLType, token, project)
				if wikiErr != nil {
					continue
				}

				projects = append(projects, wikiProject)
			}
		} else {
			break
		}
	}

	return projects
}

func getWikis(baseURL, repoURLType, token string, project Projects) (Projects, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/wikis", baseURL, project.ID)

	respBody, err := do(http.MethodGet, url, token, nil)
	if err != nil {
		slog.Error("http error",
			"error", err.Error(),
			"url", url,
		)

		return Projects{}, err
	}

	if len(respBody) < 3 {
		return Projects{}, fmt.Errorf("no found wikis")
	}

	repoURL := getRepoURL(repoURLType, project)

	return Projects{
		ID:                0,
		Name:              project.Name + ".wiki",
		PathWithNamespace: project.PathWithNamespace + ".wiki",
		HTTPURLToRepo:     strings.Replace(repoURL, ".git", ".wiki.git", 1),
	}, nil
}

func hasRegistry(baseURL, token string, project Projects) (bool, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/registry/repositories",
		baseURL,
		project.ID,
	)

	respBody, err := do(http.MethodGet, url, token, nil)
	if err != nil {
		slog.Error("http error",
			"error", err.Error(),
			"url", url,
		)

		return false, err
	}

	var registry []Registry

	err = json.Unmarshal(respBody, &registry)
	if err != nil {
		slog.Error("unmarshal error",
			"error", err.Error(),
			"url", url,
		)

		return false, err
	}

	if len(registry) == 0 {
		return false, nil
	}

	return true, nil
}

func setCleanup(baseURL, token string, project Projects) error {
	url := fmt.Sprintf("%s/api/v4/projects/%d",
		baseURL,
		project.ID,
	)

	data := []byte(`{ "container_expiration_policy_attributes": { "enabled": true, "cadence": "1d", "keep_n": "5", "older_than": "7d", "name_regex": ".*", "name_regex_keep": null } }`)

	respBody, err := do(http.MethodPut, url, token, bytes.NewBuffer(data))
	if err != nil {
		slog.Error("http error",
			"error", err.Error(),
			"url", url,
		)

		return err
	}

	_ = len(respBody) // debug

	return nil
}

func do(method, url, token string, body io.Reader) ([]byte, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	req, err := http.NewRequestWithContext(context.Background(), method, url, body)
	if err != nil {
		return nil, err
	}

	req.Header.Set(privateTokenHeader, token)
	req.Header.Set(contentTypeHeader, applicationJSONHeader)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad http response: %v", resp.Status)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return respBody, nil
}

func getRepoURL(repoURLType string, project Projects) string {
	if repoURLType == repoURLTypeHTTP {
		return project.HTTPURLToRepo
	}

	return project.SSHURLToRepo
}
