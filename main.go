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

var token string       // GitLab token for authentication
var gitlabHost string  // GitLab host URL

// Function to print usage instructions for available actions
func printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  go run main.go <action>")
	fmt.Println("Available actions:")
	fmt.Println("  versions   - Description: Creates a versions.json file with versions in stage, prod, and the last created tag in GitLab.")
	fmt.Println("  merged     - Description: Deletes branches that have already been merged.")
	fmt.Println("  renovate   - Description: Rebases all merge requests (MRs) from Renovate.")
}

func main() {

	godotenv.Load(".env") // Load environment variables from the .env file

	token = os.Getenv("GITLAB_TOKEN")     // Retrieve GitLab token
	gitlabHost = os.Getenv("GITLAB_HOST") // Retrieve GitLab host URL

	// Check if an action argument was provided
	if len(os.Args) < 2 {
		fmt.Println("Error: No action provided")
		printUsage()
		return
	}

	action := os.Args[1] // Retrieve the action argument

	// Execute the appropriate function based on the action argument
	switch action {
	case "versions":
		getVersionStatus() // Get and save version status to JSON
	case "merged":
		handlers.DeleteMergedBranches(gitlabHost, token) // Delete merged branches
	case "renovate":
		handlers.RebaseRenovateMrsForAllProjects(gitlabHost, token) // Rebase all Renovate MRs
	default:
		fmt.Println("Error: Unknown action", action)
		printUsage()
	}
}

// getVersionStatus retrieves the current versions in each environment and the latest GitLab tag, then saves them to a JSON file
func getVersionStatus() []models.VersionStatus {
	tags := handlers.GetLatestTags(gitlabHost, token) // Get the latest GitLab tags
	var versions []models.VersionStatus

	fmt.Println("versions")

	// Populate versions with deployment and GitLab tag data for each project
	for _, v := range getDeployedVersions() {
		v.GitLabTag = tags[v.Name]
		versions = append(versions, v)
	}
	saveToJsonFile("versions.json", versions) // Save the versions data to a JSON file
	return versions
}

// saveToJsonFile saves the provided object as JSON data to a specified file
func saveToJsonFile(filename string, object any) {
	content, err := json.Marshal(object) // Marshal object to JSON
	if err != nil {
		fmt.Println(err)
	}
	// Write JSON content to the specified file with read/write permissions
	err = os.WriteFile(filename, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}

// getDeployedVersions retrieves and prints the stage and prod versions for each project
func getDeployedVersions() []models.VersionStatus {
	stageVersions := handlers.GetDeployedVersionsByEnvironment("stage") // Get stage environment versions
	prodVersions := handlers.GetDeployedVersionsByEnvironment("prod")   // Get prod environment versions
	var versions []models.VersionStatus

	fmt.Println("Name\t\tStage\t\tProd")
	// Iterate through all projects, collecting stage and prod versions
	for _, project := range handlers.GetAllProjects(gitlabHost, token) {
		versions = append(versions, models.VersionStatus{
			Name:         project.Name,
			StageVersion: stageVersions[project.Name],
			ProdVersion:  prodVersions[project.Name],
		})
		fmt.Printf("%s:\t\t %s\t\t %s\n", project.Name, stageVersions[project.Name], prodVersions[project.Name])
	}
	return versions
}
