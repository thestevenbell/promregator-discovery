FROM golang:1.12.9-alpine3.10 AS build-env
LABEL stage=intermediate
WORKDIR /go/src/app
COPY . .
RUN go get -d -v ./...
RUN go install -v ./...

ENTRYPOINT ["app"]

# second stage, only include the compiled binary in the final image
FROM alpine
WORKDIR /app
COPY --from=build-env /go/bin/app .
#RUN touch ./promregator_discovery.json
#CMD chmod +rw ./promregator_discovery.json
ENTRYPOINT ["./app"]