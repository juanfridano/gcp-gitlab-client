package handlers

import (
	"context"
	"fmt"
	"log"
	"os"
	"strings"

	run "cloud.google.com/go/run/apiv2"
	"cloud.google.com/go/run/apiv2/runpb"
	"google.golang.org/api/iterator"
	"google.golang.org/api/option"
)

// GetDeployedVersionsByEnvironment retrieves deployed service versions for a specified environment in Cloud Run
func GetDeployedVersionsByEnvironment(env string) map[string]string {

	// Get the project ID from environment variables based on the specified environment (e.g., "dev", "prod")
	projectId := os.Getenv(fmt.Sprintf("PROJECT_ID_%s", strings.ToUpper(env)))
	location := "europe-west1" // Set the default location for Cloud Run services

	ctx := context.Background()

	// Initialize a new Cloud Run client using service account credentials specific to the environment
	client, err := run.NewServicesClient(ctx, option.WithCredentialsFile(fmt.Sprintf("gcp-credentials-%s.json", env)))
	if err != nil {
		log.Fatalf("Failed to create client: %s", err)
	}
	defer client.Close() // Ensure client connection is closed after function completes

	// Create a request to list all Cloud Run services in the specified project and location
	req := &runpb.ListServicesRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", projectId, location),
	}

	it := client.ListServices(ctx, req)
	fmt.Println("Cloud Run Services:")

	images := make(map[string]string) // Map to store the image name and version for each service
	for {
		resp, err := it.Next() // Iterate through the list of services
		if err == iterator.Done {
			break // Exit loop if all services are processed
		}
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			// TODO: Handle error if needed
		}

		// Extract the image name from the full container image path, ignoring the project ID prefix
		fullImageName := strings.Split(resp.Template.Containers[0].Image, fmt.Sprintf("%s/", projectId))[1]

		// TODO: Adjust image name handling for consistency across projects
		// This is specific to images starting with "dsm-rsp"; modify image name to remove prefix for normalization
		if strings.HasPrefix(fullImageName, "dsm-rsp") {
			fullImageName = strings.Split(fullImageName, "dsm-")[1]
			fmt.Printf("Image: %s\n", fullImageName)
			entry := strings.Split(fullImageName, ":")
			images[entry[0]] = entry[1] // Store the image name and version in the map
		}
	}
	fmt.Println("map:", images) // Output the map of services and their versions

	return images // Return the map containing service names and their deployed versions
}
