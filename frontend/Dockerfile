FROM node:20-alpine AS builder

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginx:alpine

RUN adduser -D -g '' appuser && \
    chown -R appuser:appuser /usr/share/nginx/html

COPY --from=builder /app/dist /usr/share/nginx/html

USER appuser

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
