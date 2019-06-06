FROM golang:alpine AS build-env
LABEL stage=intermediate
WORKDIR /go/src/app
COPY . .
VOLUME ["promregator_discovery/"]
RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["app"]

# second stage, only include the compiled binary in the final image
FROM alpine
WORKDIR /app
COPY --from=build-env /go/bin/app .
ENTRYPOINT ["./app"]