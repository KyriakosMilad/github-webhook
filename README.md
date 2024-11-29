# GitHub Webhook Handler

This Go-based zero-dep application listens for GitHub push events and processes them based on a specified branch. It allows you to handle webhooks specifically for pushes to a certain branch, such as for triggering automated deployments or other actions.

## Features

- Handles GitHub push webhooks.
- Verifies GitHub webhook signature for security.
- Responds only to pushes to a specific branch.
- Simple and easy to integrate into your existing workflow.

## Installation

1. Clone the repository:
   ```bash
   git clone https://github.com/KyriakosMilad/github-webhook.git
   cd github-webhook
    ```
2. Set up environment variables in a `.env` file (check .env.example for reference).

## Running the Application

1. Install dependencies:
   ```bash
   go mod tidy
    ```
2. Run the application:
   ```bash
   go run main.go
   ```
The server will start, and the webhook will be ready to receive push events from GitHub.

## Configuring GitHub Repo Webhook
1. Go to your GitHub repository settings.
2. Under "Webhooks", add a new webhook:
   - Payload URL: http://your-server-ip/webhook
   - Content type: application/json
   - Secret: Set it to the value of GITHUB_WEBHOOK_SECRET.
   - Select "Push" events.

## Shell script examples:

Check the `sh_examples` directory for examples of how to use the webhook handler to automate build and deployment of different application types.
Contributions to add more examples are welcome.
