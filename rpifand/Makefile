.PHONY: clean dist pkg

clean:
	rm -rf dist/bin
	rm -rf pkg/rpifand*
	rm -rf pkg/pkg
	rm -rf pkg/src

rpifand: clean
	mkdir -p dist/bin
	# https://github.com/golang/go/wiki/GoArm
	env GOOS=linux GOARCH=arm GOARM=7 \
	go build -buildmode=pie -trimpath -mod=readonly -modcacherw \
	-ldflags="-linkmode=external -extldflags=-Wl,-z,relro,-z,now" \
	-o dist/bin/rpifand cmd/rpifand/main.go

dist: rpifand
	tar -czvf pkg/rpifand.tar.gz dist/

pkg: dist
	pushd pkg && makepkg -Cc && popd
