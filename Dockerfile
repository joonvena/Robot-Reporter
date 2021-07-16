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

WORKDIR /dist

RUN cp -r /build/app /build/assets/ .

FROM alpine:latest

COPY --from=builder /dist/app /dist/assets/ /

RUN apk update && apk add --no-cache ca-certificates &&  update-ca-certificates

ENTRYPOINT ["/app"]