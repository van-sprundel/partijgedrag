FROM node:24-slim AS builder
WORKDIR /app

COPY package.json package-lock.json ./
RUN npm install --omit=dev

COPY . .

RUN npm run build

FROM node:24-slim
WORKDIR /app

ARG DATABASE_URL
ARG NODE_ENV=production
ENV DATABASE_URL=${DATABASE_URL}
ENV NODE_ENV=${NODE_ENV}

COPY --from=builder /app/backend/dist ./dist
COPY --from=builder /app/backend/node_modules ./node_modules
COPY --from=builder /app/backend/package.json ./
COPY --from=builder /app/frontend/dist ./public

EXPOSE 3000
CMD ["node", "dist/index.js"]
