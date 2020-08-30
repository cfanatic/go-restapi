# go-netchat

This module implements a REST-API interface for a [chat messenger](https://github.com/cfanatic/flutter-netchat).

## Configuration

Developed and tested on the following setup:

- macOS 10.15.6
- Go 1.15
- Docker 19.03.12
- MySQL 8.0.19

Configure the database:

1. Run a container as explained on [Docker Hub](https://hub.docker.com/_/mysql) and name it `mysql`
2. Create a database user and password in the container
3. Import the [database template](https://github.com/cfanatic/go-restapi/blob/master/scripts/database.sql) into the container
4. Define the empty [configuration parameters](https://github.com/cfanatic/go-restapi/blob/master/cmd/netchat/config.toml) for sections `[mysql]` and `[token]`

According to Wikipedia:
> In cryptography, a *salt* is random data that is used as an additional input to a one-way function that hashes data, a password or passphrase.

## Installation

Once the database is up and running, call the build process:

```bash
docker build -t netchat:latest -f Dockerfile .
docker run --name netchat -d -p 1025:1025 --link mysql:db netchat:latest
```

This yields on the console:

```bash
[...]
Step 15/17 : RUN /go/netchat/netchat -mode=init gman gman-hostname 1234
 ---> Running in 27527af0d38d
User already available
Login Hash: BHqIOuxoJeY8XPkMBpAuT87DEl8FSQI0fXOhWC5i
Removing intermediate container 27527af0d38d
 ---> cff7247e2c79
Step 16/17 : RUN /go/netchat/netchat -mode=init freeman freeman-hostname 4321
 ---> Running in 59011e1201df
User already available
Login Hash: I5QtqN5niPmNECSrkrRYotl4MQBv5kogR9BQbhWF
Removing intermediate container 59011e1201df
 ---> 7b01266c8451
Step 17/17 : ENTRYPOINT /go/netchat/netchat -mode=terminal
 ---> Running in 1cc5f8b94d07
```

Following user credentials have been initialized in the database:

| User        | Password   | Hash+Salt                                |
| ----------- |----------  | -----------------------------------------|
| gman        | 1234       | BHqIOuxoJeY8XPkMBpAuT87DEl8FSQI0fXOhWC5i |
| freeman     | 4321       | I5QtqN5niPmNECSrkrRYotl4MQBv5kogR9BQbhWF |

## Usage

Following interface handlers are implemented for use:

```Go
s.HandleFunc("/login/{user}/{password}", LoginHandler).Methods("GET")
s.HandleFunc("/user", UserHandler).Methods("GET")
s.HandleFunc("/messages/{start}/{offset}", GetMessagesHandler).Methods("GET")
s.HandleFunc("/messages/unread", GetMessagesUnreadHandler).Methods("GET")
s.HandleFunc("/message/send", SendMessageHandler).Methods("POST")
```

## Access

[https://127.0.0.1:1025/login/gman/BHqIOuxoJeY8XPkMBpAuT87DEl8FSQI0fXOhWC5i](https://127.0.0.1:1025/login/gman/BHqIOuxoJeY8XPkMBpAuT87DEl8FSQI0fXOhWC5i)

```json
{"status":"login successful"}
```

[https://127.0.0.1:1025/user](https://127.0.0.1:1025/user)

```json
{"user":"gman"}
```

[https://127.0.0.1:1025/messages/2/2](https://127.0.0.1:1025/messages/2/2)

```json
[{"name":"freeman","date":"2020-05-06 23:07:49","text":"Throw him in the mainstream."},
{"name":"gman","date":"2020-05-06 22:58:56","text":"How do you drown a hipster?"}]
```

[https://127.0.0.1:1025/messages/unread](https://127.0.0.1:1025/messages/unread)

```json
{"name":"freeman","date":"2020-05-08 18:59:10","text":":D"}
```

[https://127.0.0.1:1025/message/send](https://127.0.0.1:1025/message/send)

The message content needs to be encoded in the POST request body.

[`Future<BackendResponse> sendMessage(String text)`](https://github.com/cfanatic/flutter-netchat/blob/master/lib/backend.dart#L133) shows how this could be implemented in a messenger.
