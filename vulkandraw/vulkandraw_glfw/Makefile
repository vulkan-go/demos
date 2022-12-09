
all: windows linux darwin darwin-arm

windows:
	GOOS=windows \
	GOARCH=amd64 \
	CGO_ENABLED=1 \
	go build -ldflags="-H windowsgui"

linux:
	GOOS=linux \
	GOARCH=amd64 \
	CGO_ENABLED=1 \
	go build

darwin:
	GOOS=darwin \
	GOARCH=amd64 \
	CGO_ENABLED=1 \
	CGO_LDFLAGS="-L/usr/local/lib" \
	go build

darwin-arm:
	GOOS=darwin \
	GOARCH=arm64 \
	CGO_ENABLED=1 \
	CGO_LDFLAGS="-L/opt/homebrew/lib" \
	go build
