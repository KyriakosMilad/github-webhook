#!/bin/bash

# set yarn binary path
# this is the default linux path. change this to your yarn binary path if it's different
export PATH=$PATH:/usr/local/bin/yarn

# start the SSH agent and add the key
eval $(ssh-agent -s)
# ssh-add /root/.ssh/id_ed25519
eval $(keychain --eval --agents ssh id_ed25519)

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
yarn install || { echo "yarn install failed"; exit 1; }

# build app
echo "Building the app..."
yarn build || { echo "yarn build failed"; exit 1; }

# move production files to the production path
echo "Moving production files to the production path..."
mv dist/* /path/to/production

# restart server
echo "Restarting server..."
sudo systemctl restart nginx || { echo "Failed to restart nginx server"; exit 1; }

echo "Deployment complete!"
