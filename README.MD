This application connects to the Bluesky JetStream web socket and consumes json data for text processing. With this
application you can easily apply filters to determine what data you consume and send it downstream through built in 
text classifiers, or you can simply use the application to publish Bluesky data to an MQTT broker for consumption.

At this time the application only consumes new posts, but can be modified to fit needs easy enough.

# Getting Started

Generally, `go run main.go` to run the app, mode the .env file to change up basic configuration. For more advanced
use-cases you can use `make build` or `make docker-build` for building the application. To run the app you can use 
`make run` or `make docker-up` depending on your preference.

### Dotenvx

You will need to install a copy of dotenvx on your machine to take advantage of the .env file while using
make. [https://github.com/dotenvx/dotenvx]

### Linting

You will need a copy of golangci-lint on your machine to use the lint command in the makefile.
For more information visit [https://golangci-lint.run/welcome/install/#local-installation]

# Configuration

This table provides a detailed description of the variables defined in the `.env` file.

**General Config:**

| Variable Name | Description | Default Value | Required |
|---|---|---|---|
| DEV_MODE | Enables development mode | true | Yes |
| SERVER_PORT | Port on which the server listens | 8080 | Yes |
| JETSTREAM_URL | URL for Jetstream server | "wss://jetstream2.us-east.bsky.network/subscribe?wantedCollections=app.bsky.feed.post" | Yes |

**Custom Filtering Rules**

| Variable Name | Description | Default Value | Required |
|---|---|---|---|
| RULE_ENGLISH_ONLY | Enforces English language only posts | true | No |
| RULE_MIN_LENGTH | Enforces minimum post length | true | No |
| RULE_MIN_LENGTH_VALUE | Minimum post length (# of characters) | 10 | No |
| RULE_CONTAINS_URL | Enforces presence of a URL | true | No |
| RULE_CONTAINS_KEYWORDS | Enforces presence of specific keywords | false | No |
| RULE_CONTAINS_KEYWORDS_VALUES | List of required keywords (comma-separated, if `RULE_CONTAINS_KEYWORDS` is `true`) | false | No |
| RULE_CONTAINS_HASHTAG | Enforces presence of a hashtag | false | No |
| RULE_CONTAINS_HASHTAG_VALUES | List of required hashtags (comma-separated, if `RULE_CONTAINS_HASHTAG` is `true`) | false | No |

**ML Text Processing Flags**

| Variable Name | Description | Default Value | Required |
|---|---|---|---|
| TEXT_CATEGORY_CLASSIFIER | Enables text category classification | true | No |
| TEXT_FIN_SENTIMENT_CLASSIFIER | Enables text sentiment classification | true | No |


# Makefile Commands

This table provides a detailed description of the targets defined in the Makefile:

| Target Name | Description | Dependencies | Notes |
|---|---|---|---|
| `lint` | Lints the code using golangci-lint | | Runs `golangci-lint`
| `test` | Runs unit tests | | Runs `go test` with verbosity
| `benchmark` | Runs benchmarks | | Runs `go test` with verbosity, benchmarks, and memory profiling
| `build` | Builds the application | `lint`, `test`, `benchmark` | Builds the application after successful linting and testing
| `docker-build` | Builds the Docker image | `build` | Builds the Docker image using `docker compose build`
| `docker-up` | Starts the Docker container(s) | `docker-build` | Starts the container(s) in detached mode using `docker compose up -d`
| `docker-down` | Stops and removes the Docker container(s) | | Stops and removes container(s) using `docker compose down`
| `clean` | Cleans up build artifacts | | Removes build artifacts (`bin/*`) and attempts to remove the Docker image (`docker rmi $(IMAGE_NAME) ||

**Additional Variables:**

* `GO_VERSION`: Defines the Go version used for building.
* `BUILD_FLAGS`: Flags used during the build process.
* `VERSION`: Version information (commented out, likely retrieved dynamically).
* `GOLANGCI_LINT_PATH`: Path to the `golangci-lint` executable.
* `IMAGE_NAME`: Name of the Docker image built.
* `COMPOSE_PROFILES`: Docker Compose profile used (determined by `scripts/get_docker_profile.sh`).
