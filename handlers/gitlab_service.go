package handlers

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/juanfridano/gitlab-gcp-client/models"
)

func RebaseRenovateMrsForAllProjects(gitlabHost string, token string) {
	for _, project := range GetAllProjects(gitlabHost, token) {
		getAndRebaseRenovateMRsByProject(project, gitlabHost, token)
	}
}

func DeleteMergedBranches(gitlabHost string, token string) {

	for _, s := range GetAllProjects(gitlabHost, token) {

		url := fmt.Sprint(gitlabHost, "/api/v4/projects/", s.Id, "/repository/merged_branches")

		req, err := http.NewRequest("DELETE", url, nil)
		if err != nil {
			log.Fatal(err)
		}

		req.Header.Set("PRIVATE-TOKEN", token)
		req.Header.Set("Content-Type", "application/json")
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			log.Fatal(err)
		}
		defer resp.Body.Close()

	}
}

func deleteOldBranches(gitlabHost string, token string) {
	for _, project := range GetAllProjects(gitlabHost, token) {
		branches := getAllBranchesByProject(project, gitlabHost, token)
		fmt.Printf("Last commit authored:\t\t\tbranch:\n")
		for _, branch := range branches {

			authoredTime, err := time.Parse(time.RFC3339Nano, branch.Commit.AuthoredDate)
			if err != nil {
				fmt.Println("Could not parse time:", err)
			}

			fmt.Printf("%s\t\t%s\n", authoredTime, branch.Name)

		}
	}

}

func GetLatestTags(gitlabHost string, token string) map[string]models.GitLabTag {
	devDeploy := os.Getenv("DEPLOY_DEV_JOB")
	stageDeploy := os.Getenv("DEPLOY_STAGE_JOB")
	projects := GetAllProjects(gitlabHost, token)
	fmt.Printf("\nDEV \t STAGE \t\t Version\t\t App Name\t\t Web Url\n")

	tagsMap := make(map[string]models.GitLabTag)
	for _, project := range projects {
		url := fmt.Sprint(gitlabHost, "/api/v4/projects/", project.Id, "/repository/tags?orderBy=version&sort=desc")
		body, _ := gitlabGetRequest(url, token)
		var tags []models.Tag
		if err := json.Unmarshal([]byte(body), &tags); err != nil {
			panic(err)
		}

		if len(tags) > 0 {
			devStatus := "  ?  "
			stageStatus := "  ?  "
			latest := tags[0].Name
			webUrl := ""
			pipelineUrl := fmt.Sprint(gitlabHost, "/api/v4/projects/", project.Id, "/pipelines?scope=tags")
			body, _ := gitlabGetRequest(pipelineUrl, token)
			var pipelines []models.Pipeline
			if err := json.Unmarshal([]byte(body), &pipelines); err != nil {
				panic(err)
			}
			if len(pipelines) > 0 {
				for _, pipeline := range pipelines {
					if pipeline.Ref == latest {
						webUrl = pipeline.Link
						jobsUrl := fmt.Sprint(gitlabHost, "/api/v4/projects/", project.Id, "/pipelines/", pipeline.Id, "/jobs")
						jobsBody, _ := gitlabGetRequest(jobsUrl, token)
						var jobs []models.Job
						if err := json.Unmarshal([]byte(jobsBody), &jobs); err != nil {
							panic(err)
						}
						for _, job := range jobs {
							if job.Name == devDeploy {
								devStatus = job.Status
							}
							if job.Name == stageDeploy {
								stageStatus = job.Status
							}
						}
					}
				}

			}

			tagsMap[project.Name] = models.GitLabTag{LatestVersion: latest, DevStatus: devStatus, StageStatus: stageStatus, Link: webUrl}

			fmt.Printf("%s\t %s\t\t %s\t\t %s\t\t%s\n", devStatus, stageStatus, latest, project.Name, webUrl)
		}
	}
	return tagsMap
}

func getAndRebaseRenovateMRsByProject(project models.Project, gitlabHost string, token string) []models.MergeRequest {
	fmt.Println("Gettin branches for project: ", project.Name)

	url := fmt.Sprint(gitlabHost, "/api/v4/projects/", project.Id, "/merge_requests")
	body, _ := gitlabGetRequest(url, token)

	var mrs []models.MergeRequest
	if err := json.Unmarshal([]byte(body), &mrs); err != nil {
		panic(err)
	}

	fmt.Printf("Amount of Mrs in Project: %s , %d\n", project.Name, len(mrs))

	var renovateMRs []models.MergeRequest

	for _, s := range mrs {
		if strings.HasPrefix(s.Title, os.Getenv("RENOVATE_COMMIT_PREFIX")) {
			renovateMRs = append(renovateMRs, s)
		}

	}

	for _, s := range renovateMRs {
		fmt.Printf("Title %s iid %d \n", s.Title, s.MRId)
		rebaseMr(s, gitlabHost, token)
	}

	return mrs
}

func rebaseMr(s models.MergeRequest, gitlabHost string, token string) {

	url := fmt.Sprint(gitlabHost, "/api/v4/projects/", s.ProjectId, "/merge_requests/", s.MRId, "/rebase")
	req, err := http.NewRequest("PUT", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	fmt.Println(resp.StatusCode)
}

func getAllBranches(projects []models.Project, gitlabHost string, token string) {
	var allBranches []models.Branch
	for _, p := range projects {
		allBranches = append(allBranches, getAllBranchesByProject(p, gitlabHost, token)...)
	}
	fmt.Printf("Amount of branches: %d\n", len(allBranches))
}

func getAllBranchesByProject(project models.Project, gitlabHost string, token string) []models.Branch {
	fmt.Println("Getting branches for project: ", project.Name)

	url := fmt.Sprint(gitlabHost, "/api/v4/projects/", project.Id, "/repository/branches")
	body, _ := gitlabGetRequest(url, token)

	var branches []models.Branch
	if err := json.Unmarshal([]byte(body), &branches); err != nil {
		panic(err)
	}

	fmt.Printf("Amount of branches in Project: %s , %d\n", project.Name, len(branches))

	return branches
}

func GetAllProjects(gitlabHost string, token string) []models.Project {
	groups := strings.Split(os.Getenv("GROUP_IDS"), " ")

	var allProjects []models.Project

	for _, s := range groups {
		projects, _ := getProjectsInGroup(s, gitlabHost, token)
		allProjects = append(allProjects, projects...)
	}
	return allProjects
}

func getProjectsInGroup(groupId string, gitlabHost string, token string) ([]models.Project, error) {
	fmt.Println("Retrieving projets in group id", groupId)

	url := fmt.Sprint(gitlabHost, "/api/v4/groups/", groupId, "/projects")

	body, err := gitlabGetRequest(url, token)

	var projects []models.Project
	if err := json.Unmarshal([]byte(body), &projects); err != nil {
		panic(err)
	}

	return projects, err
}

func gitlabGetRequest(url string, token string) ([]byte, error) {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		log.Fatal(err)
	}

	req.Header.Set("PRIVATE-TOKEN", token)
	req.Header.Set("Content-Type", "application/json")
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}
	// fmt.Printf("Request to %s with status %v \n", url, resp.StatusCode)
	return body, err
}
