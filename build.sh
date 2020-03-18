VERSION=$(git describe --always --tags --long)
go build -ldflags "-X github.com/ontio/bonus/config.Version=${VERSION}"  -o bonus main.go
#CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -ldflags "-X github.com/ontio/bonus/config.Version=${VERSION}"  -o bonus main.go
#CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -ldflags "-X github.com/ontio/bonus/config.Version=${VERSION}" main.go
