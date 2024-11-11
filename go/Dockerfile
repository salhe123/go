# Dockerfile for Go backend

# Build the Go app
FROM golang:1.23.2 AS Build

WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN go build -o main .

# Run the Go app
FROM golang:1.23.2

WORKDIR /app
COPY --from=build /app/main .
EXPOSE 7070
CMD ["./main"]