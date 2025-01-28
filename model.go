package main

const (
	repoURLTypeHTTP = "http"
)

type Projects struct {
	ID                int    `json:"id"`
	Name              string `json:"name"`
	PathWithNamespace string `json:"path_with_namespace"`
	SSHURLToRepo      string `json:"ssh_url_to_repo"`
	HTTPURLToRepo     string `json:"http_url_to_repo"`
}

type Registry struct {
	ID       int    `json:"id"`
	Path     string `json:"path"`
	Location string `json:"location"`
}
