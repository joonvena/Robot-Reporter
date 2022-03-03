FROM golang:alpine AS builder

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=amd64

WORKDIR /build

COPY go.mod .
COPY go.sum .
RUN go mod download

COPY . .

RUN go build -o app .

FROM alpine:latest

COPY --from=builder /build/app .
COPY  assets /assets

RUN apk update && apk add --no-cache ca-certificates &&  update-ca-certificates

ENTRYPOINT ["/app"]