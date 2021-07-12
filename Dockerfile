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

RUN CGO_ENABLED=0 go build -o app .

WORKDIR /dist

RUN cp /build/app /build/template.txt .

FROM alpine:latest

COPY --from=builder /dist/app /dist/template.txt /

RUN apk update && apk add --no-cache ca-certificates &&  update-ca-certificates && apk add --no-cache libc6-compat

ENTRYPOINT ["/app"]