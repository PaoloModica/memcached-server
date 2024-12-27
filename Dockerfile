FROM golang:1.21 AS build-stage

WORKDIR /memcached

COPY go.mod ./
RUN go mod download

COPY internal/store/*.go ./internal/store/
COPY internal/app/*.go ./internal/app/
COPY cmd/memcached/main.go ./cmd/memcached/

RUN CGO_ENABLED=0 GOOS=linux go build -o ccmemcached cmd/memcached/main.go

FROM build-stage AS build-test-stage
WORKDIR /memcached
RUN go test -v ./...

FROM gcr.io/distroless/base-debian11 AS build-release-stage

EXPOSE ${PORT}

WORKDIR /memcached

COPY --from=build-stage /memcached/ccmemcached ./ccmemcached

ENTRYPOINT ["./ccmemcached"]
