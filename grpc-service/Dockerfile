# syntax=docker/dockerfile:1

FROM golang:1.16-alpine

WORKDIR /app

COPY go.mod ./
COPY go.sum ./

RUN go mod download

COPY . .

RUN apk add build-base

RUN go build /app/cmd/main.go

EXPOSE 50051 

CMD [ "./main" ]

