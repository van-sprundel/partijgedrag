FROM node:24-slim AS base
WORKDIR /app

COPY package.json package-lock.json* ./
COPY backend/package.json ./backend/
COPY frontend/package.json ./frontend/

FROM base AS dependencies
RUN npm ci

FROM node:24-slim AS build
WORKDIR /app

COPY --from=dependencies /app/node_modules ./node_modules
COPY . .
RUN npm run db:generate --prefix backend
RUN npm run build

FROM node:24-slim
WORKDIR /app

ARG DATABASE_URL
ENV DATABASE_URL=${DATABASE_URL}
ENV NODE_ENV=production

COPY --from=base /app/ .
RUN npm ci --production

COPY --from=build /app/backend/dist ./dist
COPY --from=build /app/frontend/dist ./public
COPY --from=build /app/backend/prisma ./prisma
COPY --from=build /app/node_modules/@prisma/client ./node_modules/@prisma/client

EXPOSE 3000
CMD ["node", "dist/index.js"]
