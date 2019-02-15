GOOS=linux GOARCH=amd64 go build -o lambda-main main.go
zip lambda-main.zip lambda-main