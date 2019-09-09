FROM golang:1.12.9-alpine3.10 AS build-env
LABEL stage=intermediate
RUN apk add git
WORKDIR /go/src/app
COPY . .
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux go build -o app
ENTRYPOINT ["app"]


# second stage, only include the compiled binary in the final image
FROM alpine:3.10.1

EXPOSE ${PORT:-8080}

WORKDIR /root/
COPY --from=build-env /app .

ENTRYPOINT ["./app"]
