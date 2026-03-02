# Tracker-only Dockerfile - watching, tally dumps, optional on-chain updates
# Dashboard has its own Dockerfile.dashboard - rebuild dashboard separately
#
# Build: docker build -t snapshot-activity-tracker .

FROM golang:1.25-alpine AS tracker-builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o snapshot-activity-tracker .

FROM golang:1.25-alpine AS tracker
WORKDIR /app
RUN apk add --no-cache ca-certificates curl
COPY --from=tracker-builder /app/snapshot-activity-tracker .
COPY --from=tracker-builder /app/contract/ ./contract/
EXPOSE 8001
CMD ["./snapshot-activity-tracker"]
