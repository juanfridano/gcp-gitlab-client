package models

type Project struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
}

type Branch struct {
	Name   string `json:"name"`
	Merged bool   `json:"merged"`
	Commit Commit `json:"commit"`
}

type Commit struct {
	AuthoredDate  string `json:"authored_date"`
	CommittedDate string `json:"committed_date"`
}

type MergeRequest struct {
	Id           int    `json:"id"`
	MRId         int    `json:"iid"`
	ProjectId    int    `json:"project_id"`
	AddLAbels    string `json:"add_labels"`
	Title        string `json:"title"`
	SourceBranch string `json:"source_branch"`
	TargetBranch string `json:"target_branch"`
}

type Tag struct {
	Name string `json:"name"`
}

type Pipeline struct {
	Id   int    `json:"id"`
	Name string `json:"name"`
	Ref  string `json:"ref"`
	Link string `json:"web_url"`
}

type Job struct {
	Id     int    `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
	Stage  string `json:"stage"`
}

type GitLabTag struct {
	LatestVersion string `json:"latest_version"`
	DevStatus     string `json:"dev_status"`
	StageStatus   string `json:"stage_status"`
	Link          string `json:"link"`
}
