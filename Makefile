GO ?= go

.PHONY: overlord
overlord:
	$(GO) build -o overlord ./main.go

.PHONY: dylib
dylib:
	CGO_ENABLED=1 $(GO) build -buildmode=c-shared -o overlord.dylib ./dylib/

.PHONY: sliver
sliver:
	GOOS=darwin GOARCH=amd64 $(GO) build -o ./sliver/bin/macos/overlord-amd64 ./main.go
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GO) build -buildmode=c-shared -o ./sliver/bin/macos/overlord-amd64.dylib ./dylib/

	GOOS=windows GOARCH=amd64 $(GO) build -o ./sliver/bin/windows/overlord-amd64.exe -ldflags -H=windowsgui ./main.go

	GOOS=linux GOARCH=amd64 $(GO) build -o ./sliver/bin/linux/overlord-amd64 ./main.go

clean:
	rm -f overlord overlord.dylib
