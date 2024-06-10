FROM golang:alpine AS builder

# inject the target architecture (https://docs.docker.com/reference/dockerfile/#automatic-platform-args-in-the-global-scope)
ARG TARGETARCH

ENV GO111MODULE=on \
    CGO_ENABLED=0 \
    GOOS=linux \
    GOARCH=${TARGETARCH}

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
