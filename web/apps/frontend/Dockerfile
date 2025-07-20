# ---- Base Node Image ----
FROM node:20-alpine AS base
WORKDIR /app

# ---- Install dev dependencies (with caching) ----
FROM base AS development-dependencies-env
COPY package.json package-lock.json ./
RUN npm ci
# Only now copy rest of app
COPY . .

# ---- Install prod dependencies ----
FROM base AS production-dependencies-env
COPY package.json package-lock.json ./
RUN npm ci --omit=dev

# ---- Build app (uses dev deps from previous stage) ----
FROM base AS build-env
COPY --from=development-dependencies-env /app /app
RUN npm run build --workspaces --if-present

# ---- Final runtime image ----
FROM base AS final
COPY package.json package-lock.json ./
COPY --from=production-dependencies-env /app/node_modules ./node_modules
COPY --from=build-env /app/build ./build

# Set NODE_ENV to production
ENV NODE_ENV=production

CMD ["npm", "run", "start"]