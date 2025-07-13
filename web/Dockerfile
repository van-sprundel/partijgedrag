# Use an official Node.js image as a base
FROM node:20-alpine AS builder

# Set working directory
WORKDIR /app

# Copy package.json and package-lock.json first to leverage caching
COPY package.json package-lock.json ./

# Copy workspaces' package.json files
COPY apps/backend/package.json apps/backend/
COPY apps/frontend/package.json apps/frontend/
COPY apps/frontend/server.js apps/frontend/
COPY packages/ts-rest-contracts/package.json packages/ts-rest-contracts/

# Install dependencies (including workspaces)
RUN npm ci --workspaces

# Copy the rest of the project files
COPY . .

# Build all workspaces
RUN npm run build

# Create a production image
FROM node:20-alpine AS runner

# Set working directory
WORKDIR /app

# Copy the node_modules and built files from the builder stage
COPY --from=builder /app/node_modules ./node_modules
COPY --from=builder /app/apps ./apps
COPY --from=builder /app/packages ./packages
COPY --from=builder /app/package.json ./package.json

# Expose ports (adjust as needed)
EXPOSE 3000

# Start the application (modify based on how you run both frontend & backend)
CMD ["npm", "run", "serve"]