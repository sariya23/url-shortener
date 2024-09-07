# docker run --net=host -it --rm server:v1

FROM golang:1.23.1-alpine3.19 as base
FROM base as dev
WORKDIR /app
COPY . .
RUN go mod download
RUN go build /app/cmd/url-shortner/main.go