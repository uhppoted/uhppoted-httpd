RELEASE = v0.8.x
DIST   ?= development
DEBUG  ?= --debug
CMD     = ./bin/uhppoted-httpd

.DEFAULT_GOAL := test
.PHONY: sass
.PHONY: debug
.PHONY: reset
.PHONY: update
.PHONY: vet
.PHONY: lint
.PHONY: vuln
.PHONY: update-release
.PHONY: quickstart

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
	# go get -u github.com/hyperjumptech/grule-rule-engine
	go mod tidy

update-release:
	go get -u github.com/uhppoted/uhppote-core
	go get -u github.com/uhppoted/uhppoted-lib
	go mod tidy

update-all:
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
	mkdir -p httpd/html/images/default
	sass --no-source-map sass/themes/light:httpd/html/css/default
	sass --no-source-map sass/themes/light:httpd/html/css/light
	sass --no-source-map sass/themes/dark:httpd/html/css/dark
	cp httpd/html/images/light/* httpd/html/images/default
	npx eslint --fix httpd/html/javascript/*.js
	go build -trimpath -o bin/ ./...

test: build
	go test -tags "tests" ./...

vet: 
	go vet ./...

lint: 
	env GOOS=darwin  GOARCH=amd64 staticcheck ./...
	env GOOS=linux   GOARCH=amd64 staticcheck ./...
	env GOOS=windows GOARCH=amd64 staticcheck ./...
	npx eslint httpd/html/javascript/*.js

vuln:
	govulncheck ./...

benchmark: 
	go build -trimpath -o bin ./...
	# go test -bench=. ./...
	go test -count 5 -bench=.  ./system/events

coverage: build
	go test -cover ./...

build-all: test vet lint
	mkdir -p dist/$(DIST)/windows
	mkdir -p dist/$(DIST)/darwin
	mkdir -p dist/$(DIST)/linux
	mkdir -p dist/$(DIST)/arm
	mkdir -p dist/$(DIST)/arm7
	env GOOS=linux   GOARCH=amd64         GOWORK=off go build -trimpath -o dist/$(DIST)/linux   ./...
	env GOOS=linux   GOARCH=arm64         GOWORK=off go build -trimpath -o dist/$(DIST)/arm     ./...
	env GOOS=linux   GOARCH=arm   GOARM=7 GOWORK=off go build -trimpath -o dist/$(DIST)/arm7    ./...
	env GOOS=darwin  GOARCH=amd64         GOWORK=off go build -trimpath -o dist/$(DIST)/darwin  ./...
	env GOOS=windows GOARCH=amd64         GOWORK=off go build -trimpath -o dist/$(DIST)/windows ./...
	cp -r httpd/html dist/$(DIST)

build-quickstart: 
	cp -r httpd/html documentation/starter-kit/etc/httpd/html
	tar --directory=documentation/starter-kit --exclude=".DS_Store" -cvzf ./dist/quickstart-darwin_$(RELEASE).tar.gz  . -C ../../dist/$(DIST)/darwin  .
	tar --directory=documentation/starter-kit --exclude=".DS_Store" -cvzf ./dist/quickstart-linux_$(RELEASE).tar.gz   . -C ../../dist/$(DIST)/linux   .
	tar --directory=documentation/starter-kit --exclude=".DS_Store" -cvzf ./dist/quickstart-windows_$(RELEASE).tar.gz . -C ../../dist/$(DIST)/windows .
	tar --directory=documentation/starter-kit --exclude=".DS_Store" -cvzf ./dist/quickstart-arm7_$(RELEASE).tar.gz    . -C ../../dist/$(DIST)/arm7    .

release: update-release build-all
	find . -name ".DS_Store" -delete
	tar --directory=dist --exclude=".DS_Store" -cvzf dist/$(DIST).tar.gz $(DIST)
	cd dist;  zip --recurse-paths $(DIST).zip $(DIST)

publish: release
	echo "Releasing version $(VERSION)"
	rm -f dist/development.tar.gz
	gh release create "$(VERSION)" "./dist/uhppoted-httpd_$(VERSION).tar.gz" "./dist/uhppoted-httpd_$(VERSION).zip" --draft --prerelease --title "$(VERSION)-beta" --notes-file release-notes.md

debug: format
	go build -trimpath -o bin ./...
	go test -tags "tests" -run TestTimezoneGMT2 ./types
	# mkdir -p dist/test
	# tar xvzf dist/quickstart-darwin.tar.gz --directory dist/test
	# cd dist/test && ./uhppoted-httpd --debug --console --config uhppoted.conf

delve: format
	go build -trimpath -o bin ./...
#	dlv exec ./bin/uhppoted-httpd -- --debug --console
	dlv test github.com/uhppoted/uhppoted-httpd/system/interfaces -- run TestLANSet

# NTS: 1. sass --watch doesn't seem to consistently pick up changes in themed partials
#      2. For development only - doesn't build the default CSS because the duplication 
#         of light and default creates a naming conflict if run in the same command
#         i.e. find sass -name "*.scss" | entr sass --no-source-map sass/stylesheets:html/css/default sass/themes/light:html/css/light sass/themes/dark:html/css/dark
sass:
	find sass -name "*.scss" | entr sass --no-source-map sass/themes/light:httpd/html/css/light sass/themes/dark:httpd/html/css/dark

godoc:
	godoc -http=:80	-index_interval=60s

version: build
	$(CMD) version

help: build
	$(CMD) help
	$(CMD) help commands
	$(CMD) help version
	$(CMD) help help

run: build
	$(CMD) --debug --console

monitor: build
	$(CMD) --debug --console --mode monitor

synchronize: build
	$(CMD) --debug --console --mode synchronize

daemonize: build
	sudo $(CMD) daemonize

undaemonize: build
	sudo $(CMD) undaemonize

config: build
	$(CMD) config

