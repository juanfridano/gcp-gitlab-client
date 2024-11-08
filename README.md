
# GitLab GCP Client

This application provides utilities for managing GitLab projects and deployments with a focus on version control, merged branch cleanup, and Renovate merge request (MR) handling. The tool leverages GitLab and Google Cloud Run APIs to automate these operations across environments.

## Table of Contents

- [Getting Started](#getting-started)
- [Environment Setup](#environment-setup)
- [Usage](#usage)
- [Actions](#actions)
- [Example Output](#example-output)
- [License](#license)

## Getting Started

### Prerequisites

- [Go](https://golang.org/doc/install) (version specified in `go.mod`)
- Git
- GitLab credentials (access token and host)
- GCP Credentials (environment-specific JSON files, such as `gcp-credentials-[ENV].json`)

### Installation

1. Clone the repository:
   ```sh
   git clone https://github.com/yourusername/gitlab-gcp-client.git
   cd gitlab-gcp-client
   ```

2. Install dependencies:
   ```sh
   go mod download
   ```

## Environment Setup

1. Duplicate the `.env-example` file and rename it to `.env`:
   ```sh
   cp .env-example .env
   ```

2. Open the `.env` file and add your GitLab credentials and environment-specific project IDs:

   **Example of `.env` variables**:
   ```plaintext
   GITLAB_TOKEN=your_gitlab_access_token
   GITLAB_HOST=https://gitlab.yourdomain.com
   PROJECT_ID_STAGE=your_stage_project_id
   PROJECT_ID_PROD=your_prod_project_id
   GROUP_IDS="group_id1 group_id2"
   DEPLOY_DEV_JOB=your_deployment_job_name
   DEPLOY_STAGE_JOB=your_stage_job_name
   RENOVATE_COMMIT_PREFIX=renovate
   ```

These variables configure connections to your GitLab instance and specify project IDs for different environments.

3. Store GCP credentials for each environment in the repository root, following the naming format `gcp-credentials-[ENV].json`, replacing `[ENV]` with the environment (e.g., `stage`, `prod`).

## Usage

Run the application using the following command:

```sh
go run main.go <action>
```

Replace `<action>` with one of the supported actions below.

## Actions

### 1. `versions`
Generates a `versions.json` file with deployment details for each GitLab project, including the deployed versions in staging and production, along with the latest GitLab tag.

```sh
go run main.go versions
```

**Description**: Creates a `versions.json` file, structured as follows:
- **`name`**: Project name.
- **`prod_version`**: Current production version.
- **`stage_version`**: Current staging version.
- **`gitlab_tag`**: Contains the latest tag, deployment status for development and staging, and a link to the pipeline.

**Sample Output**:
```json
[
    {
        "name": "project-name",
        "prod_version": "1.0.0",
        "stage_version": "1.1.0",
        "gitlab_tag": {
            "latest_version": "1.1.0",
            "dev_status": "success",
            "stage_status": "success",
            "link": "https://gitlab.example.com/project-name/-/pipelines/12345"
        }
    }
]
```

### 2. `merged`
Deletes branches that have already been merged into the main branch in GitLab, helping to maintain a clean repository.

```sh
go run main.go merged
```

**Description**: This action iterates over all projects, identifying and removing merged branches from GitLab.

### 3. `renovate`
Rebases all merge requests created by Renovate across your GitLab projects to keep them up-to-date with the target branch.

```sh
go run main.go renovate
```

**Description**: This action fetches open Renovate MRs and rebases them to ensure compatibility with the latest changes in the target branch.

## Example Output

Running the `versions` action will produce a `versions.json` file, structured like the example below:

```json
[
    {
        "name": "jfr-evaluator",
        "prod_version": "",
        "stage_version": "0.5.0",
        "gitlab_tag": {
            "latest_version": "0.5.0",
            "dev_status": "success",
            "stage_status": "success",
            "link": "https://gitlab.example.com/jfr-evaluator/-/pipelines/2199186"
        }
    }
]
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
