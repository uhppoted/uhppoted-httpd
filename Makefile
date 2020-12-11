VERSION = v0.7.x
LDFLAGS = -ldflags "-X uhppote.VERSION=$(VERSION)" 
DIST   ?= development
DEBUG  ?= --debug
CMD     = ./bin/uhppoted-httpd

.PHONY: debug
.PHONY: reset
.PHONY: bump
.PHONY: build-all

all: test      \
	 benchmark \
     coverage

clean:
	go clean
	rm -rf bin

format: 
	go fmt ./...

build: format
	mkdir -p bin
	go build -o bin ./...
	mkdir -p html/images/default
	sass --no-source-map html/sass/themes/light:html/css/default
	sass --no-source-map html/sass/themes/light:html/css/light
	sass --no-source-map html/sass/themes/dark:html/css/dark
	cp html/images/light/* html/images/default
	npx eslint --fix html/javascript/*.js

test: build
	go test ./...

vet: build
	go vet ./...

lint: build
	golint ./...
	npx eslint html/javascript/*.js

benchmark: build
	go test -bench ./...

coverage: build
	go test -cover ./...

build-all: 
	mkdir -p bin
	go build -o bin ./...
	go test ./...
	go vet ./...
	mkdir -p dist/$(DIST)/windows
	mkdir -p dist/$(DIST)/darwin
	mkdir -p dist/$(DIST)/linux
	mkdir -p dist/$(DIST)/arm7
	env GOOS=linux   GOARCH=amd64       go build -o dist/$(DIST)/linux   ./...
	env GOOS=linux   GOARCH=arm GOARM=7 go build -o dist/$(DIST)/arm7    ./...
	env GOOS=darwin  GOARCH=amd64       go build -o dist/$(DIST)/darwin  ./...
	env GOOS=windows GOARCH=amd64       go build -o dist/$(DIST)/windows ./...

release: build-all
	find . -name ".DS_Store" -delete
	tar --directory=dist --exclude=".DS_Store" -cvzf dist/$(DIST).tar.gz $(DIST)
	cd dist; zip --recurse-paths $(DIST).zip $(DIST)

bump:
	go get -u github.com/uhppoted/uhppote-core
	go get -u github.com/uhppoted/uhppoted-api
	go get -u github.com/cristalhq/jwt/v3
	go get -u github.com/google/uuid
	go get -u golang.org/x/sys

debug: build
#	go test -v ./... -run "TestCardAdd"
	cp html/images/light/* html/images/default
	find html/sass -name "*.scss" | entr sass --no-source-map html/sass/themes/light:html/css/default

# sass --watch doesn't seem to consistently pick up changes in themed partials
sass:
	find html/sass -name "*.scss" | entr sass --no-source-map html/sass/stylesheets:html/css/default html/sass/themes/light:html/css/light html/sass/themes/dark:html/css/dark

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
	$(CMD) --console

reset:
	cp /usr/local/var/com.github.uhppoted/httpd/memdb/db.orig /usr/local/var/com.github.uhppoted/httpd/memdb/db.json
