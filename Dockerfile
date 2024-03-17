FROM golang:1.21.4-alpine3.18 AS build

RUN apk --no-cache add ca-certificates

WORKDIR /zsv-playground

COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags '-w -s'

FROM ubuntu:22.04

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/
COPY --from=build zsv-playground .

EXPOSE 8080

CMD ["./zsv-playground"]
