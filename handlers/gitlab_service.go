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

// RebaseRenovateMrsForAllProjects initiates a rebase operation for all Renovate merge requests (MRs) across all projects
func RebaseRenovateMrsForAllProjects(gitlabHost string, token string) {
	for _, project := range GetAllProjects(gitlabHost, token) {
		getAndRebaseRenovateMRsByProject(project, gitlabHost, token)
	}
}

// DeleteMergedBranches deletes all merged branches across all projects
func DeleteMergedBranches(gitlabHost string, token string) {
	for _, s := range GetAllProjects(gitlabHost, token) {
		url := fmt.Sprint(gitlabHost, "/api/v4/projects/", s.Id, "/repository/merged_branches")

		// Create a DELETE request to remove merged branches
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
		defer resp.Body.Close() // Ensure the response body is closed
	}
}

// deleteOldBranches lists branches and their last commit dates to identify outdated branches for each project
func deleteOldBranches(gitlabHost string, token string) {
	for _, project := range GetAllProjects(gitlabHost, token) {
		branches := getAllBranchesByProject(project, gitlabHost, token)
		fmt.Printf("Last commit authored:\t\t\tbranch:\n")
		for _, branch := range branches {

			// Parse the branch's last commit authored date
			authoredTime, err := time.Parse(time.RFC3339Nano, branch.Commit.AuthoredDate)
			if err != nil {
				fmt.Println("Could not parse time:", err)
			}

			fmt.Printf("%s\t\t%s\n", authoredTime, branch.Name)
		}
	}
}

// GetLatestTags retrieves the latest tags and deployment status for each project, outputting results to the console
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

		// Check if tags exist for the project and retrieve the latest one
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

			// Check pipeline jobs for the latest tag's deployment statuses
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

// getAndRebaseRenovateMRsByProject retrieves and rebases Renovate MRs for a given project
func getAndRebaseRenovateMRsByProject(project models.Project, gitlabHost string, token string) []models.MergeRequest {
	fmt.Println("Getting branches for project: ", project.Name)

	url := fmt.Sprint(gitlabHost, "/api/v4/projects/", project.Id, "/merge_requests")
	body, _ := gitlabGetRequest(url, token)

	var mrs []models.MergeRequest
	if err := json.Unmarshal([]byte(body), &mrs); err != nil {
		panic(err)
	}

	fmt.Printf("Amount of MRs in Project: %s , %d\n", project.Name, len(mrs))

	// Filter MRs with Renovate commit prefix
	var renovateMRs []models.MergeRequest
	for _, s := range mrs {
		if strings.HasPrefix(s.Title, os.Getenv("RENOVATE_COMMIT_PREFIX")) {
			renovateMRs = append(renovateMRs, s)
		}
	}

	// Rebase each Renovate MR
	for _, s := range renovateMRs {
		fmt.Printf("Title %s iid %d \n", s.Title, s.MRId)
		rebaseMr(s, gitlabHost, token)
	}

	return mrs
}

// rebaseMr sends a PUT request to rebase a specific MR
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

	fmt.Println(resp.StatusCode) // Output the response status code
}

// getAllBranches retrieves all branches for a list of projects and outputs the total count
func getAllBranches(projects []models.Project, gitlabHost string, token string) {
	var allBranches []models.Branch
	for _, p := range projects {
		allBranches = append(allBranches, getAllBranchesByProject(p, gitlabHost, token)...)
	}
	fmt.Printf("Amount of branches: %d\n", len(allBranches))
}

// getAllBranchesByProject retrieves all branches for a specific project
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

// GetAllProjects retrieves all projects within specified GitLab groups (defined in environment variables)
func GetAllProjects(gitlabHost string, token string) []models.Project {
	groups := strings.Split(os.Getenv("GROUP_IDS"), " ")

	var allProjects []models.Project
	for _, s := range groups {
		projects, _ := getProjectsInGroup(s, gitlabHost, token)
		allProjects = append(allProjects, projects...)
	}
	return allProjects
}

// getProjectsInGroup retrieves projects for a specific GitLab group by its ID
func getProjectsInGroup(groupId string, gitlabHost string, token string) ([]models.Project, error) {
	fmt.Println("Retrieving projects in group id", groupId)

	url := fmt.Sprint(gitlabHost, "/api/v4/groups/", groupId, "/projects")
	body, err := gitlabGetRequest(url, token)

	var projects []models.Project
	if err := json.Unmarshal([]byte(body), &projects); err != nil {
		panic(err)
	}

	return projects, err
}

// gitlabGetRequest sends a GET request to the GitLab API and returns the response body as a byte array
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
