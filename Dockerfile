# -----------------------------------------------------------------------------
# Build...
FROM golang:1.12.3-alpine AS build

ENV VERSION=v0.3.0 GO111MODULE=on PACKAGE=github.com/derailed/popeye

WORKDIR /go/src/$PACKAGE

COPY go.mod go.sum main.go ./
COPY internal internal
COPY pkg pkg
COPY cmd cmd

RUN apk --no-cache add git ;\
  CGO_ENABLED=0 GOOS=linux go build -o /go/bin/popeye \
  -ldflags="-w -s -X $PACKAGE/cmd.version=$VERSION" *.go


# -----------------------------------------------------------------------------
# Image...
FROM alpine:3.9.3
COPY --from=build /go/bin/popeye /bin/popeye
ENTRYPOINT [ "/bin/popeye" ]