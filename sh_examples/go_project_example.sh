#!/bin/bash

# define project directory path
PROJECT_DIR="/path/to/your/project"

# change to the project directory
echo "Changing to project directory..."
cd "$PROJECT_DIR" || { echo "Directory not found: $PROJECT_DIR"; exit 1; }

# pull latest changes
echo "Pulling latest changes..."
git pull

# install dependencies
echo "Installing dependencies..."
go mod tidy

# build app
echo "Building the app..."
go build -o backendApp

# move production files to production path
echo "Moving production files to the production path..."
mv backendApp /path/to/production

# restart server
echo "Restarting server..."
sudo systemctl restart backend

echo "Deployment complete!"