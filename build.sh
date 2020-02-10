VERSION=$(git describe --always --tags --long)
go build -ldflags "-X github.com/ontio/bonus/config.Version=${VERSION}"  -o bonus main.go
