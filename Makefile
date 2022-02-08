VERSION = v0.7.x
LDFLAGS = -ldflags "-X uhppote.VERSION=$(VERSION)" 
DIST   ?= development
DEBUG  ?= --debug
CMD     = ./bin/uhppoted-httpd

.PHONY: debug
.PHONY: reset
.PHONY: update
.PHONY: update-release

all: test      \
	 benchmark \
     coverage

clean:
	go clean
	rm -rf bin

update:
	go get -u github.com/uhppoted/uhppote-core@master
	go get -u github.com/uhppoted/uhppoted-lib@master
	go get -u github.com/cristalhq/jwt/v3
	go get -u github.com/google/uuid
	go get -u golang.org/x/sys
	go get -u github.com/hyperjumptech/grule-rule-engine
	go mod tidy

update-release:
	go get -u github.com/uhppoted/uhppote-core
	go get -u github.com/uhppoted/uhppoted-lib
	go get -u github.com/cristalhq/jwt/v3
	go get -u github.com/google/uuid
	go get -u golang.org/x/sys
	go get -u github.com/hyperjumptech/grule-rule-engine
	go mod tidy

format: 
	go fmt ./...

build: format
	mkdir -p bin
	go build -trimpath -o bin ./...
	mkdir -p html/images/default
	sass --no-source-map html/sass/themes/light:html/css/default
	sass --no-source-map html/sass/themes/light:html/css/light
	sass --no-source-map html/sass/themes/dark:html/css/dark
	cp html/images/light/* html/images/default
	npx eslint --fix html/javascript/*.js

test: build
	go test ./...

vet: test
	go vet ./...

lint: vet
	golint ./...
	npx eslint html/javascript/*.js

benchmark: build
	go test -bench ./...

coverage: build
	go test -cover ./...

build-all: vet
	mkdir -p dist/$(DIST)/windows
	mkdir -p dist/$(DIST)/darwin
	mkdir -p dist/$(DIST)/linux
	mkdir -p dist/$(DIST)/arm7
	env GOOS=linux   GOARCH=amd64       go build -trimpath -o dist/$(DIST)/linux   ./...
	env GOOS=linux   GOARCH=arm GOARM=7 go build -trimpath -o dist/$(DIST)/arm7    ./...
	env GOOS=darwin  GOARCH=amd64       go build -trimpath -o dist/$(DIST)/darwin  ./...
	env GOOS=windows GOARCH=amd64       go build -trimpath -o dist/$(DIST)/windows ./...

release: update-release build-all
	find . -name ".DS_Store" -delete
	tar --directory=dist --exclude=".DS_Store" -cvzf dist/$(DIST).tar.gz $(DIST)
	cd dist; zip --recurse-paths $(DIST).zip $(DIST)

debug: format
	go build -trimpath -o bin ./...
	go test -v -run Test ./auth/...
	# dlv test github.com/uhppoted/uhppoted-httpd/system/catalog

# NOTE: sass --watch doesn't seem to consistently pick up changes in themed partials
sass:
	find html/sass -name "*.scss" | entr sass --no-source-map html/sass/themes/light:html/css/light html/sass/themes/dark:html/css/dark
	# find html/sass -name "*.scss" | entr sass --no-source-map html/sass/stylesheets:html/css/default html/sass/themes/light:html/css/light html/sass/themes/dark:html/css/dark

version: build
	$(CMD) version

help: build
	$(CMD) help
	$(CMD) help commands
	$(CMD) help version
	$(CMD) help help

daemonize: build
	sudo $(CMD) daemonize

undaemonize: build
	sudo $(CMD) undaemonize

run: build
	# $(CMD) --debug --console > ../runtime/debug.log
	$(CMD) --debug --console

delve: build
	dlv exec ./bin/uhppoted-httpd -- --debug --console
