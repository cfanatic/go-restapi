compile:
	go build -o bin/netchat cmd/netchat/main.go

init:
	@go build -o bin/netchat cmd/netchat/main.go
	@./bin/netchat -mode init

deploy:
	@go build -o bin/netchat cmd/netchat/main.go
	@./bin/netchat -mode terminal

clean:
	rm -f -r bin/netchat
