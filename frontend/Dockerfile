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

RUN adduser -D -g '' appuser
USER appuser

EXPOSE 80

CMD ["nginx", "-g", "daemon off;"]
