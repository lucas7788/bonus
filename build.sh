VERSION=$(git describe --always --tags --long)
CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/ontio/bonus/config.Version=${VERSION}"  -o bonus main.go
