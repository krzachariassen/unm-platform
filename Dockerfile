# Stage 1: Build frontend
FROM node:22-alpine AS frontend-build
WORKDIR /app/frontend
COPY frontend/package.json ./
RUN npm install --no-package-lock
COPY frontend/ ./
RUN npm run build

# Stage 2: Build backend
FROM golang:1.25-alpine AS backend-build
WORKDIR /app/backend
COPY backend/ ./
RUN CGO_ENABLED=0 GOOS=linux go build -mod=vendor -o /server ./cmd/server/

# Stage 3: Minimal runtime
FROM alpine:3.21
RUN apk add --no-cache ca-certificates tzdata
WORKDIR /app

COPY --from=backend-build /server /app/server
COPY --from=frontend-build /app/frontend/dist /app/dist
COPY config/ /app/config/

ENV UNM_ENV=production

EXPOSE 8080

ENTRYPOINT ["/app/server"]
