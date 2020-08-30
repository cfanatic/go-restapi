FROM golang:alpine AS build
WORKDIR /go/netchat
COPY . .
RUN go build -o bin/netchat cmd/netchat/main.go

FROM alpine:latest
RUN apk add --no-cache nano
RUN apk add --no-cache tzdata
RUN apk add --no-cache openssl
ENV TZ Europe/Berlin
WORKDIR /go/netchat
COPY --from=build /go/netchat/bin .
COPY --from=build /go/netchat/cmd/netchat misc/
EXPOSE 1025
RUN openssl req -x509 -nodes -days 365 \
    -subj "/C=DE/ST=BW/O=codefanatic/CN=codefanatic.de" \
    -newkey rsa:2048 -keyout misc/server.key \
    -out misc/server.crt
RUN /go/netchat/netchat -mode=init gman gman-hostname 1234
RUN /go/netchat/netchat -mode=init freeman freeman-hostname 4321
ENTRYPOINT /go/netchat/netchat -mode=terminal
