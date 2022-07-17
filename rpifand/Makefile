.PHONY: clean bins dist pkg_aarch64

clean:
	rm -rf dist/aarch64
	rm -rf pkg/aarch64/rpifand*
	rm -rf pkg/aarch64/pkg
	rm -rf pkg/aarch64/src

bins: clean
	mkdir -p dist/aarch64/
	# https://github.com/golang/go/wiki/GoArm
	env GOOS=linux GOARCH=arm64 \
	go build -trimpath -o dist/aarch64/rpifand cmd/rpifand/main.go

dist: bins
	tar -czvf pkg/aarch64/rpifand.tar.gz dist/

pkg_aarch64: dist
	pushd pkg/aarch64 && CARCH=aarch64 makepkg -Cc && popd
