GO ?= go

.PHONY=overlord
overlord:
	$(GO) build -o overlord ./main.go

.PHONY=dylib
dylib:
	CGO_ENABLED=1 $(GO) build -buildmode=c-shared -o overlord.dylib ./dylib/

clean:
	rm -f overlord
