FROM node:20-alpine AS builder

ARG VITE_API_BASE_URL=http://localhost:8080

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN VITE_API_BASE_URL=$VITE_API_BASE_URL npm run build

FROM nginx:alpine

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

RUN adduser -D -g '' appuser && chown -R appuser:appuser /usr/share/nginx/html
USER appuser

EXPOSE 80

HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost/health || exit 1

CMD ["nginx", "-g", "daemon off;"]
