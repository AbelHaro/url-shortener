FROM node:20-alpine AS builder

ARG VITE_API_BASE_URL

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN VITE_API_BASE_URL=$VITE_API_BASE_URL npm run build

FROM golang:1.24-alpine AS server

RUN go install github.com/johanbrandhorst/serve@latest

FROM scratch

COPY --from=builder /app/dist /www
COPY --from=server /go/bin/serve /serve

EXPOSE 80

CMD ["/serve", "-addr", ":80", "-dir", "/www"]
