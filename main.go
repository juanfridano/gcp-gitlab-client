package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"

	"github.com/joho/godotenv"
	"github.com/juanfridano/gitlab-gcp-client/handlers"
	"github.com/juanfridano/gitlab-gcp-client/models"
)

var token string

var gitlabHost string

// Function to print usage instructions
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  go run main.go <action>")
	fmt.Println("Available actions:")
	fmt.Println("  versions   - Description: Creats a versions.json. JSON includes versions in stage, prod and last created tag in gitlab.")
	fmt.Println("  merged   - Description: Deletes branches if already merged.")
	fmt.Println("  renovate - Description: Rebases all MRs from Renovate")
}

func main() {

	godotenv.Load(".env")

	token = os.Getenv("GITLAB_TOKEN")
	gitlabHost = os.Getenv("GITLAB_HOST")

	if len(os.Args) < 2 {
		fmt.Println("Error: No action provided")
		printUsage()
		return
	}

	action := os.Args[1]

	switch action {
	case "versions":
		getVersionStatus()
	case "merged":
		handlers.DeleteMergedBranches(gitlabHost, token)
	case "renovate":
		handlers.RebaseRenovateMrsForAllProjects(gitlabHost, token)
	default:
		fmt.Println("Error: Unknown action", action)
		printUsage()
	}
}

func getVersionStatus() []models.VersionStatus {
	tags := handlers.GetLatestTags(gitlabHost, token)
	var versions []models.VersionStatus

	fmt.Println("versions")

	for _, v := range getDeployedVersions() {
		v.GitLabTag = tags[v.Name]
		versions = append(versions, v)
	}
	saveToJsonFile("versions.json", versions)
	return versions
}

func saveToJsonFile(filename string, object any) {

	content, err := json.Marshal(object)
	if err != nil {
		fmt.Println(err)
	}
	err = os.WriteFile(filename, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

func getDeployedVersions() []models.VersionStatus {

	stageVersions := handlers.GetDeployedVersionsByEnvironment("stage")
	prodVersions := handlers.GetDeployedVersionsByEnvironment("prod")
	var versions []models.VersionStatus

	fmt.Println("Name\t\tStage\t\tProd")
	for _, project := range handlers.GetAllProjects(gitlabHost, token) {
		versions = append(versions, models.VersionStatus{Name: project.Name, StageVersion: stageVersions[project.Name], ProdVersion: prodVersions[project.Name]})
		fmt.Printf("%s:\t\t %s\t\t %s\n", project.Name, stageVersions[project.Name], prodVersions[project.Name])
	}
	return versions
}
