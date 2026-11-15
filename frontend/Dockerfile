# syntax=docker/dockerfile:1

FROM node:22-alpine AS build

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci --no-audit --no-fund

COPY . .

ARG BUILD_CONFIGURATION=production
ARG BUILD_SHA=unknown
ARG FRONTEND_API_URL

RUN node -e "const fs = require('fs'); const path = 'src/environments/environment.ts'; let apiUrl = process.env.FRONTEND_API_URL; if (!apiUrl) { throw new Error('FRONTEND_API_URL build arg is required'); } apiUrl = apiUrl.replace(/\\/+$/, ''); if (!apiUrl.endsWith('/api/v1')) apiUrl += '/api/v1'; new URL(apiUrl); const marker = \"'__FRONTEND_API_URL__'\"; const source = fs.readFileSync(path, 'utf8'); if (!source.includes(marker)) { throw new Error('FRONTEND_API_URL placeholder not found'); } fs.writeFileSync(path, source.replace(marker, JSON.stringify(apiUrl)));"
RUN npx ng build --configuration=${BUILD_CONFIGURATION}

FROM nginx:1.27-alpine

ARG BUILD_SHA=unknown
LABEL org.opencontainers.image.revision=$BUILD_SHA

RUN rm /etc/nginx/conf.d/default.conf

COPY nginx.conf /etc/nginx/conf.d/default.conf
COPY --from=build /app/dist/tex-talk/browser /usr/share/nginx/html

RUN printf '%s\n' "$BUILD_SHA" > /usr/share/nginx/html/build-sha.txt \
  && addgroup -S appgroup \
  && adduser -S appuser -G appgroup \
  && chown -R appuser:appgroup /usr/share/nginx/html \
  && chown -R appuser:appgroup /var/cache/nginx \
  && chown -R appuser:appgroup /var/log/nginx \
  && touch /var/run/nginx.pid \
  && chown appuser:appgroup /var/run/nginx.pid

USER appuser

EXPOSE 8080

CMD ["nginx", "-g", "daemon off;"]
