FROM golang:1.21-alpine AS dev
RUN apk --update add curl gcc git musl-dev util-linux-dev

WORKDIR /app

COPY go.* ./
RUN go mod download
RUN go mod verify

COPY assets assets
COPY tools tools
RUN go generate tools/javascript.go

COPY cmd cmd
COPY internal internal
COPY views views

FROM dev AS build
RUN mkdir -p bin
ENV CGO_ENABLED=0
ENV GOOS=linux
ENV GOARCH=amd64
RUN go build -trimpath -mod=readonly -ldflags "-extldflags '-static' -s -w" -o bin ./cmd/...

RUN chmod -R a+x /app/bin

FROM alpine AS release
RUN apk --update add --no-cache ca-certificates curl tzdata && update-ca-certificates
RUN addgroup -S appuser && adduser -DHS appuser appuser

COPY --chown=appuser:appuser --from=build /app/bin /app

USER appuser
WORKDIR /app

FROM release AS web
CMD ["./web"]

FROM release AS load-prices
CMD ["./loadprices"]

FROM release AS load-securities
CMD ["./loadsecurities"]
