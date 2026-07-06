.PHONY: build run test hash clean

build:
	go build -o hugo-cms ./cmd/server

run:
	go run ./cmd/server

test:
	go test ./...

hash:
	@read -p "Password: " pwd && go run ./cmd/hashpassword "$$pwd"

clean:
	rm -f hugo-cms
