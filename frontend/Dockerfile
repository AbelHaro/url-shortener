FROM node:20-alpine AS builder

ARG VITE_API_BASE_URL

WORKDIR /app

COPY package*.json ./
RUN npm ci

COPY . .
RUN VITE_API_BASE_URL=$VITE_API_BASE_URL npm run build

FROM nginx:alpine

RUN adduser -D -g '' appuser && \
    chown -R appuser:appuser /usr/share/nginx/html

COPY --from=builder /app/dist /usr/share/nginx/html
COPY nginx.conf /etc/nginx/conf.d/default.conf

USER appuser

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
