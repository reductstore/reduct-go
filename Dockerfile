FROM golang:1.24.2-bookworm

WORKDIR /app

COPY go.mod .
COPY go.sum .

# copy vendor folder if it exists - to support private repos
COPY vendor/ vendor/

RUN go mod vendor

COPY . .

RUN go build -o ./main

EXPOSE 8080

CMD ["./main"]

