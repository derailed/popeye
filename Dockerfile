# -----------------------------------------------------------------------------
# Build...
FROM golang:1.23-alpine3.19 AS build
ARG TARGETOS
ARG TARGETARCH

WORKDIR /popeye

COPY go.mod go.sum main.go Makefile ./
COPY internal internal
COPY cmd cmd
COPY types types
COPY pkg pkg
RUN apk --no-cache add make git gcc libc-dev curl ca-certificates binutils-gold && \
    CGO_ENABLED=0 GOOS=${TARGETOS:-linux} GOARCH=${TARGETARCH} make build

# -----------------------------------------------------------------------------
# Image...
FROM alpine:3.19.1

COPY --from=build /etc/ssl/certs/ca-certificates.crt /etc/ssl/certs/ca-certificates.crt
COPY --from=build /popeye/execs/popeye /bin/popeye

RUN adduser -u 5000 -D nonroot
USER 5000

ENTRYPOINT [ "/bin/popeye" ]