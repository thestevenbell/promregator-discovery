FROM golang:alpine AS build-env
LABEL stage=intermediate
RUN apk add git
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["app"]

# second stage, only include the compiled binary in the final image
FROM alpine
WORKDIR /app
COPY --from=build-env /go/bin/app .
EXPOSE ${PORT:-8080}
ENTRYPOINT ["./app"]