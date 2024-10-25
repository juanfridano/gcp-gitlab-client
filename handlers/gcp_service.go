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

func GetDeployedVersionsByEnvironment(env string) map[string]string {

	projectId := os.Getenv(fmt.Sprintf("PROJECT_ID_%s", strings.ToUpper(env)))
	location := "europe-west1"

	ctx := context.Background()

	client, err := run.NewServicesClient(ctx, option.WithCredentialsFile(fmt.Sprintf("gcp-credentials-%s.json", env)))
	if err != nil {
		log.Fatalf("Failed to create client: %s", err)
	}

	defer client.Close()

	req := &runpb.ListServicesRequest{
		Parent: fmt.Sprintf("projects/%s/locations/%s", projectId, location),
	}

	it := client.ListServices(ctx, req)
	fmt.Println("Cloud Run Services:")

	images := make(map[string]string)
	for {
		resp, err := it.Next()
		if err == iterator.Done {
			break
		}
		if err != nil {
			fmt.Printf("Error: %s\n", err)
			// TODO: Handle error.
		}

		fullImageName := strings.Split(resp.Template.Containers[0].Image, fmt.Sprintf("%s/", projectId))[1]

		// TODO: this is project specific, find a way of normalizing image/container naes and git repository names
		if strings.HasPrefix(fullImageName, "dsm-rsp") {
			fullImageName = strings.Split(fullImageName, "dsm-")[1]
			fmt.Printf("Image: %s\n", fullImageName)
			entry := strings.Split(fullImageName, ":")
			images[entry[0]] = entry[1]
		}
	}
	fmt.Println("map:", images)

	return images
}
