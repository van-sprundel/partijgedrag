FROM node:24-slim AS base
WORKDIR /app
COPY package.json package-lock.json* ./
COPY backend/package.json ./backend/
COPY frontend/package.json ./frontend/

FROM base AS dependencies
RUN npm ci

FROM node:24-slim AS backend-build
WORKDIR /app
COPY --from=dependencies /app/node_modules ./node_modules
COPY backend/ ./backend/
RUN npm run db:generate --prefix backend
RUN npm run build --prefix backend

FROM node:24-slim AS frontend-build
WORKDIR /app
COPY --from=dependencies /app/node_modules ./node_modules
COPY backend/ ./backend/
COPY frontend/ ./frontend/
RUN npm run db:generate --prefix backend
RUN npm run build --prefix frontend

FROM node:24-slim
WORKDIR /app

ARG DATABASE_URL
ARG NODE_ENV=production
ENV DATABASE_URL=${DATABASE_URL}
ENV NODE_ENV=${NODE_ENV}

COPY package.json package-lock.json* ./
COPY backend/package.json ./backend/
RUN npm ci --production --ignore-scripts

COPY --from=backend-build /app/backend/dist ./dist
COPY --from=frontend-build /app/frontend/dist ./public
COPY --from=backend-build /app/backend/prisma ./prisma

COPY --from=backend-build /app/node_modules/.prisma ./node_modules/.prisma

EXPOSE 3000
CMD ["node", "dist/index.js"]
