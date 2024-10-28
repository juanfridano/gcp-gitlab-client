
# GitLab GCP Client

This application provides utilities for interacting with GitLab to manage project versions, merged branches, and Renovate merge requests (MRs). The tool uses environment variables to configure access to your GitLab instance.

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
- GCP Credentials (access Key) saved pro environment in the format "gcp-credentials-[ENV].json"

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

2. Open the `.env` file and add your GitLab credentials:

   **Example** of `.env` variables:
   ```
   GITLAB_TOKEN=your_gitlab_access_token
   GITLAB_HOST=https://gitlab.yourdomain.com
   ```

These environment variables are required to connect to your GitLab instance and perform API requests.

## Usage

Run the application using the following command:

```sh
go run main.go <action>
```

Replace `<action>` with one of the supported actions below.

## Actions

### 1. `versions`
Generates a `versions.json` file with version details for all GitLab projects, including the deployed versions in staging and production, as well as the latest GitLab tag.

```sh
go run main.go versions
```

**Description**: Creates a `versions.json` file, structured as follows:
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
    },
    ...
]
```

### 2. `merged`
Deletes branches that have already been merged into the main branch in GitLab.

```sh
go run main.go merged
```

**Description**: This action removes any merged branches, helping to keep your GitLab repository clean and manageable.

### 3. `renovate`
Rebases all merge requests created by Renovate across your GitLab projects.

```sh
go run main.go renovate
```

**Description**: Rebases open Renovate merge requests to bring them up to date with the target branch.

## Example Output

Running the `versions` action will produce a `versions.json` file structured like the example below:

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
    },
    ...
]
```

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.
