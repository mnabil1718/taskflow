FROM node:24-slim AS deps
WORKDIR /app
COPY package.json package-lock.json ./
# npm install, not ci: the committed lockfile omits two platform-optional
# nested deps (@emnapi/*) that npm refuses to record on this host, which makes
# `npm ci` reject the lock. install tolerates it and still honors the pinned
# versions already in the lock.
RUN npm install

FROM node:24-slim AS builder
WORKDIR /app
COPY --from=deps /app/node_modules ./node_modules
COPY . .
ARG NEXT_PUBLIC_API_URL=http://localhost:8080/api/v1
ENV NEXT_PUBLIC_API_URL=$NEXT_PUBLIC_API_URL
RUN npm run build

FROM node:24-slim AS runner
WORKDIR /app
ENV NODE_ENV=production
COPY --from=builder /app/public ./public
COPY --from=builder /app/.next/standalone ./
COPY --from=builder /app/.next/static ./.next/static
EXPOSE 3000
CMD ["node", "server.js"]
