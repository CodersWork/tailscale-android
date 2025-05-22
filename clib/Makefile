GOOS = linux
GOARCH = arm64
OUT = output/libts.a

all:
	CC=/root/ohos-cli/sdk/default/openharmony/native/llvm/bin/aarch64-unknown-linux-ohos-clang \
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=1 CGO_CFLAGS="-I." \
	/root/golang_go/bin/go build -v -tlsmodegd -buildmode=c-archive -o $(OUT)

clean:
	rm -rf output
