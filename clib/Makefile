CGO_ENABLED=1
GOOS = linux
GOARCH = arm64
OUT = output/libts.a

all:
	GOOS=$(GOOS) GOARCH=$(GOARCH) CGO_ENABLED=$(CGO_ENABLED) CC=aarch64-linux-gnu-gcc \
	CGO_CFLAGS="-I." go build -v -buildmode=c-archive -o $(OUT)

clean:
	rm -rf output
