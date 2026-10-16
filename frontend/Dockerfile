FROM node:24-alpine AS build
WORKDIR /app
COPY package*.json ./
RUN npm ci
COPY tsconfig*.json vite.config.ts index.html ./
COPY src ./src
COPY public ./public
RUN npm run build

FROM nginx:stable-alpine
COPY --from=build /app/dist /usr/share/nginx/html
RUN mkdir -p /etc/nginx/templates && printf '%s\n' \
    'server {' \
    '    listen 80;' \
    '    root /usr/share/nginx/html;' \
    '    index index.html;' \
    '    location /api/ {' \
    '        proxy_pass ${API_UPSTREAM}/;' \
    '    }' \
    '    location / {' \
    '        try_files $uri $uri/ /index.html;' \
    '    }' \
    '}' > /etc/nginx/templates/default.conf.template
ENV API_UPSTREAM=http://host.docker.internal:8080
EXPOSE 80
