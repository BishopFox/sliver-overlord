GO ?= go

.PHONY=overlord
overlord:
	$(GO) build -o overlord ./main.go

.PHONY=lib
lib:
	CGO_ENABLED=1 $(GO) build -buildmode=c-shared -o overlord.dylib ./example/

clean:
	rm -f overlord
