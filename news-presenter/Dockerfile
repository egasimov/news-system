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

RUN go build -o /news-presenter-go-svc -v ./cmd/main.go
#RUN ls -la

##
## Deploy
##
#FROM golang:1.19.1-buster
FROM gcr.io/distroless/base-debian10

WORKDIR /

COPY --from=build /news-presenter-go-svc /app/news-presenter-go-svc-exec
#COPY --from=build /app/config/config.json /app/config/config.json

#RUN ls /app/
#RUN ls /app/config/
#RUN cat /app/config/config.json
EXPOSE 8081

USER root

WORKDIR /app

ENTRYPOINT ["./news-presenter-go-svc-exec"]
