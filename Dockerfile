FROM golang:1.12.0-stretch AS build-env
LABEL stage=intermediate
COPY . /app
WORKDIR /app
ENV GO111MODULE=on
RUN CGO_ENABLED=0 GOOS=linux go build -o app

# second stage, only include the compiled binary in the final image
FROM alpine:3.10.1
WORKDIR /root/
COPY --from=build-env /app .
ENTRYPOINT ["./app"]