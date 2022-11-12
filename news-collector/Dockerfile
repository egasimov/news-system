# syntax=docker/dockerfile:1

##
## Build
##
FROM golang:1.19-buster AS build

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . .

RUN go build -o /news-collector-go-svc -v ./cmd/main.go
#RUN ls -la

##
## Deploy
##
#FROM golang:1.19.1-buster
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /news-collector-go-svc /app/news-collector-go-svc-exec
COPY --from=build /app/config/config.json /app/config/config.json

#RUN ls /app/
#RUN ls /app/config/
#RUN cat /app/config/config.json
EXPOSE 8080

USER root

WORKDIR /app

ENTRYPOINT ["./news-collector-go-svc-exec"]
