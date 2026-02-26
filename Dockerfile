# Stage 1: Build snapshot-activity-tracker
FROM golang:1.25-alpine AS tracker-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o snapshot-activity-tracker .

# Stage 2: Build dashboard-api
FROM golang:1.25-alpine AS dashboard-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o dashboard-api ./cmd/dashboard-api

# Stage 3: Build frontend
FROM node:20-alpine AS frontend-builder
WORKDIR /app
COPY frontend/package*.json ./
RUN npm install
COPY frontend/ ./
RUN npm run build

# Stage 4: Final runtime - snapshot-activity-tracker (default target)
FROM golang:1.25-alpine
WORKDIR /app
RUN apk add --no-cache ca-certificates curl
COPY --from=tracker-builder /app/snapshot-activity-tracker .
EXPOSE 8001
CMD ["./snapshot-activity-tracker"]

# Stage 5: Dashboard runtime
FROM golang:1.25-alpine AS dashboard
WORKDIR /app
RUN apk add --no-cache ca-certificates wget
# Copy dashboard-api binary
COPY --from=dashboard-builder /app/dashboard-api .
# Copy frontend build
COPY --from=frontend-builder /app/dist ./frontend/dist
EXPOSE 8080
CMD ["./dashboard-api"]
