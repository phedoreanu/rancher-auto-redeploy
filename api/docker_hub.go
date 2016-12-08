package api

// DockerHubRequest mimics DockerHub's request.
type DockerHubRequest struct {
	CallbackURL string      `json:"callback_url"`
	PushData    *PushData   `json:"push_data"`
	Repository  *Repository `json:"repository"`
}

// PushData mimics DockerHub's JSON model.
type PushData struct {
	Tag string
}

// Repository mimics DockerHub's JSON model.
type Repository struct {
	Name     string
	RepoName string `json:"repo_name"`
}
