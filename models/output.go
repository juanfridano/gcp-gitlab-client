package models

type VersionStatus struct {
	Name         string    `json:"name"`
	ProdVersion  string    `json:"prod_version"`
	StageVersion string    `json:"stage_version"`
	GitLabTag    GitLabTag `json:"gitlab_tag"`
}

type LogEntry struct {
	severity string `json:"severity"`
	message  string `json:"message"`
}
