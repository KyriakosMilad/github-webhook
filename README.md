# GitHub Webhook Handler

This Go-based application listens for GitHub push events and processes them based on a specified branch. It
allows you to handle webhooks specifically for pushes to a certain branch, can be used for triggering automated deployments
or other actions.

## Features

- Handles GitHub push webhooks.
- Verifies GitHub webhook signature for security.
- Responds only to pushes to a specific branch.
- Simple and easy to integrate into your existing workflow.

## Prerequisites

- Golang installed.

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

## Starting the Application on Server Start

Ensure the GitHub Webhook handler starts automatically when the server boots.

1. **Create a systemd service file** (e.g., `github-webhook.service`):
   ```shell
   touch github-webhook.service
   ```

   ```ini
   [Unit]
   Description=GitHub Webhook Handler
   After=network.target

   [Service]
   ExecStart=/path/to/go/bin run /path/to/this/repository/on/your/server/main.go
   WorkingDirectory=/path/to/this/repository/on/your/server
   EnvironmentFile=/path/to/this/repository/on/your/server/.env
   Restart=always
   User=your-user
   Group=your-user

   [Install]
   WantedBy=multi-user.target
    ```
   Replace ``your-user`` with the user that will run the service (make sure user has the permissions to read/write to your app-repo path), and the paths with the actual paths on your server.


2. Copy the service file to the systemd directory:
   ```bash
   sudo cp github-webhook.service /etc/systemd/system/
    ```
3. Enable and start the service:
   ```bash
   sudo systemctl enable github-webhook
   sudo systemctl start github-webhook
    ```

## Shell script examples:

Check the `sh_examples` directory for examples of how to use the webhook handler to automate build and deployment of
different application types. Make sure you give the shell script the permissions to do its stuff, also make sure to make it executable.
Contributions to add more examples are welcome.
