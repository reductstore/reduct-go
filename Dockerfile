FROM golang:1.24.2-bookworm

WORKDIR /app

COPY go.mod .
COPY go.sum .

RUN go mod download

COPY . .

# Run tests
CMD ["go", "test", "-v", "./..."]

