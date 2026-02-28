FROM node:20-alpine AS builder

ARG VITE_API_BASE_URL

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN VITE_API_BASE_URL=$VITE_API_BASE_URL npm run build

FROM caddy:2-alpine

COPY --from=builder /app/dist /usr/share/caddy

EXPOSE 80