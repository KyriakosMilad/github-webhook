#!/bin/bash

# define project directory path
PROJECT_DIR="/path/to/your/project"

# change to the project directory
echo "Changing to project directory..."
cd "$PROJECT_DIR" || { echo "Directory not found: $PROJECT_DIR"; exit 1; }

# pull latest changes
echo "Pulling latest changes..."
git pull || { echo "Git pull failed"; exit 1; }

# install dependencies
echo "Installing dependencies..."
go mod tidy || { echo "go mod tidy failed"; exit 1; }

# build app
echo "Building the app..."
go build -o backendApp || { echo "Go build failed"; exit 1; }

# move production files to production path
echo "Moving production files to the production path..."
mv backendApp /path/to/production || { echo "Move failed"; exit 1; }

# restart server
echo "Restarting server..."
sudo systemctl restart backend || { echo "Failed to restart backend service"; exit 1; }

echo "Deployment complete!"