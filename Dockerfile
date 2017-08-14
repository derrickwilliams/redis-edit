FROM golang:1.8

WORKDIR /go/src/github.com/derrickwilliams/redis-edit
ADD . .

CMD go run main.go