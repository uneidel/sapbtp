VERSION="1.0"

clean:
	rm ./build/*

build-cli: 
	cd src && CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -o ../build/sapbtp-amd64 -ldflags='-X main.Version=$(VERSION) -extldflags "-static"' 
	cd src && CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -a -o ../build/sapbtp-darwin -ldflags='-X main.Version=$(VERSION) -extldflags "-static"' 
	cd src && CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -a -o ../build/sapbtp-win.exe -ldflags='-X main.Version=$(VERSION) -extldflags "-static"' 
	