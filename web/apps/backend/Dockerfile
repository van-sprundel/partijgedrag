# Use a Node.js base image
# Base image
FROM node:18 AS base

WORKDIR /app

# Copy package.json and package-lock.json to install dependencies
COPY package.json package-lock.json ./
COPY packages ./packages
COPY apps/backend ./apps/backend

# Install dependencies using npm workspaces
RUN npm install

# Build step if needed (e.g., TypeScript)
RUN npm run build --workspace=apps/backend

# Set the command to start the backend
CMD ["npm", "run", "start", "--workspace=apps/backend"]
