FROM golang:1.24.2-bookworm

WORKDIR /app

RUN go mod download

COPY . .

# Run tests 
CMD ["go", "test", "-v", "./..."]

