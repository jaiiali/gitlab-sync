package main

import (
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
	pageSize           = 20
	privateTokenHeader = "PRIVATE-TOKEN"
)

func getProjects(baseURL, token, orderBy string) []Projects {
	var (
		pageNumber int
		projects   []Projects
	)

	for {
		pageNumber++
		url := fmt.Sprintf("%s/api/v4/projects?order_by=%s&sort=desc&per_page=%d&page=%d",
			baseURL,
			orderBy,
			pageSize,
			pageNumber,
		)

		body, err := do(url, token)
		if err != nil {
			slog.Error("http error",
				"error", err.Error(),
				"url", url,
			)

			continue
		}

		var currentProjects []Projects

		err = json.Unmarshal(body, &currentProjects)
		if err != nil {
			slog.Error("unmarshal error",
				"error", err.Error(),
				"url", url,
			)

			continue
		}

		if len(currentProjects) > 0 {
			projects = append(projects, currentProjects...)

			for _, project := range currentProjects {
				wikiProject, wikiErr := getWikis(baseURL, token, project)
				if wikiErr != nil {
					continue
				}

				projects = append(projects, wikiProject)
			}
		} else {
			break
		}

		// time.Sleep(3 * time.Second)
	}

	return projects
}

func getWikis(baseURL, token string, project Projects) (Projects, error) {
	url := fmt.Sprintf("%s/api/v4/projects/%d/wikis", baseURL, project.ID)

	body, err := do(url, token)
	if err != nil {
		slog.Error("http error",
			"error", err.Error(),
			"url", url,
		)

		return Projects{}, err
	}

	if len(body) < 3 {
		return Projects{}, fmt.Errorf("no found wikis")
	}

	return Projects{
		ID:                0,
		Name:              project.Name + ".wiki",
		PathWithNamespace: project.PathWithNamespace + ".wiki",
		SSHURLToRepo:      strings.Replace(project.SSHURLToRepo, ".git", ".wiki.git", 1),
	}, nil
}

func do(url, token string) ([]byte, error) {
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}

	client := http.Client{
		Transport: transport,
		Timeout:   15 * time.Second,
	}

	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set(privateTokenHeader, token)

	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("bad http response: %v", resp.Status)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	err = resp.Body.Close()
	if err != nil {
		return nil, err
	}

	return body, nil
}
