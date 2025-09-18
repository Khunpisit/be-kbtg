# Build stage
FROM golang:1.22-alpine AS build
WORKDIR /app
RUN apk add --no-cache build-base
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o server .

# Runtime stage
FROM alpine:3.20
RUN apk add --no-cache ca-certificates sqlite-libs
WORKDIR /app
COPY --from=build /app/server ./
ENV DB_PATH=/data/app.db
VOLUME ["/data"]
EXPOSE 3000
CMD ["./server"]