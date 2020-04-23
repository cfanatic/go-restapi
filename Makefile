compile:
	go build -o bin/netchat cmd/netchat/main.go

deploy:
	@go build -o bin/netchat cmd/netchat/main.go
	@./bin/netchat

clean:
	rm -f -r bin/netchat
