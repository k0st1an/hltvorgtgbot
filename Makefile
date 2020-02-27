darwin:
	GOOS=darwin GOARCH=amd64 go build -o hltvorgbot-darwin-amd64 .

linux:
	GOOS=linux GOARCH=amd64 go build -o hltvorgbot-linux-amd64 .

.PHONY: darwin linux
