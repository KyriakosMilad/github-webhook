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
yarn install

# build app
echo "Building the app..."
yarn build

# move production files to production path
echo "Moving production files to the production path..."
mv dist/* /path/to/production

# restart server
echo "Restarting server..."
sudo systemctl restart nginx

echo "Deployment complete!"