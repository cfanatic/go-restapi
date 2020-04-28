FROM golang:alpine AS build
WORKDIR /go/netchat
COPY . .
RUN go build -o bin/netchat cmd/netchat/main.go

FROM alpine:latest
RUN apk add --no-cache nano
RUN apk add --no-cache tzdata
ENV TZ Europe/Berlin
WORKDIR /go/netchat
COPY --from=build /go/netchat/bin .
COPY --from=build /go/netchat/misc misc/
EXPOSE 1025
ENTRYPOINT /go/netchat/netchat
