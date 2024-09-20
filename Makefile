RELEASE = v0.8.x
DIST   ?= development
DEBUG  ?= --debug
CMD     = ./bin/uhppoted-httpd
DOCKER  ?= ghcr.io/uhppoted/httpd:latest

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
	go get -u github.com/uhppoted/uhppote-core@main
	go get -u github.com/uhppoted/uhppoted-lib@main
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

build-all: build test vet lint
	mkdir -p dist/$(DIST)/linux
	mkdir -p dist/$(DIST)/arm
	mkdir -p dist/$(DIST)/arm7
	mkdir -p dist/$(DIST)/darwin-x64
	mkdir -p dist/$(DIST)/darwin-arm64
	mkdir -p dist/$(DIST)/windows
	env GOOS=linux   GOARCH=amd64         GOWORK=off go build -trimpath -o dist/$(DIST)/linux        ./...
	env GOOS=linux   GOARCH=arm64         GOWORK=off go build -trimpath -o dist/$(DIST)/arm          ./...
	env GOOS=linux   GOARCH=arm   GOARM=7 GOWORK=off go build -trimpath -o dist/$(DIST)/arm7         ./...
	env GOOS=darwin  GOARCH=amd64         GOWORK=off go build -trimpath -o dist/$(DIST)/darwin-x64   ./...
	env GOOS=darwin  GOARCH=arm64         GOWORK=off go build -trimpath -o dist/$(DIST)/darwin-arm64 ./...
	env GOOS=windows GOARCH=amd64         GOWORK=off go build -trimpath -o dist/$(DIST)/windows      ./...
	cp -r httpd/html dist/$(DIST)

build-quickstart: 
	cp -r httpd/html documentation/starter-kit/etc/httpd/html
	tar --directory=documentation/starter-kit --exclude=".DS_Store" -cvzf ./dist/quickstart-darwin_$(RELEASE).tar.gz  . -C ../../dist/$(DIST)/darwin  .
	tar --directory=documentation/starter-kit --exclude=".DS_Store" -cvzf ./dist/quickstart-linux_$(RELEASE).tar.gz   . -C ../../dist/$(DIST)/linux   .
	tar --directory=documentation/starter-kit --exclude=".DS_Store" -cvzf ./dist/quickstart-windows_$(RELEASE).tar.gz . -C ../../dist/$(DIST)/windows .
	tar --directory=documentation/starter-kit --exclude=".DS_Store" -cvzf ./dist/quickstart-arm7_$(RELEASE).tar.gz    . -C ../../dist/$(DIST)/arm7    .

release: update-release build-all
	find . -name ".DS_Store" -delete
	tar --directory=dist/$(DIST)/linux        --exclude=".DS_Store" -cvzf dist/$(DIST)-linux-x64.tar.gz    .
	tar --directory=dist/$(DIST)/arm          --exclude=".DS_Store" -cvzf dist/$(DIST)-arm-x64.tar.gz      .
	tar --directory=dist/$(DIST)/arm7         --exclude=".DS_Store" -cvzf dist/$(DIST)-arm7.tar.gz         .
	tar --directory=dist/$(DIST)/darwin-x64   --exclude=".DS_Store" -cvzf dist/$(DIST)-darwin-x64.tar.gz   .
	tar --directory=dist/$(DIST)/darwin-arm64 --exclude=".DS_Store" -cvzf dist/$(DIST)-darwin-arm64.tar.gz .
	cd dist/$(DIST)/windows && zip --recurse-paths ../../$(DIST)-windows-x64.zip . -x ".DS_Store"

publish: release
	echo "Releasing version $(VERSION)"
	gh release create "$(VERSION)" "./dist/$(DIST)-arm-x64.tar.gz"      \
	                               "./dist/$(DIST)-arm7.tar.gz"         \
	                               "./dist/$(DIST)-darwin-arm64.tar.gz" \
	                               "./dist/$(DIST)-darwin-x64.tar.gz"   \
	                               "./dist/$(DIST)-linux-x64.tar.gz"    \
	                               "./dist/$(DIST)-windows-x64.zip"     \
	                               --draft --prerelease --title "$(VERSION)-beta" --notes-file release-notes.md

debug: build
	$(CMD) --debug --console

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

docker: docker-dev docker-ghcr docker-compose
	cd docker && find . -name .DS_Store -delete && rm -f compose.zip && zip --recurse-paths compose.zip compose
	
docker-dev: build
	rm -rf dist/docker/dev/*
	mkdir -p dist/docker/dev
	mkdir -p dist/docker/dev/system
	mkdir -p dist/docker/dev/grules
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o dist/docker/dev ./...
	cp docker/dev/Dockerfile    dist/docker/dev
	cp docker/dev/uhppoted.conf dist/docker/dev
	cp docker/dev/auth.json     dist/docker/dev
	cp docker/dev/acl.grl       dist/docker/dev
	cp -r docker/dev/grules     dist/docker/dev
	cp -r docker/dev/system     dist/docker/dev
	cd dist/docker/dev && docker build --no-cache -f Dockerfile -t uhppoted/uhppoted-httpd-dev .

docker-ghcr: build
	rm -rf dist/docker/ghcr/*
	mkdir -p dist/docker/ghcr
	env GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -trimpath -o dist/docker/ghcr ./...
	cp docker/ghcr/Dockerfile    dist/docker/ghcr
	cp docker/ghcr/uhppoted.conf dist/docker/ghcr
	cp docker/ghcr/auth.json     dist/docker/ghcr
	cp docker/ghcr/acl.grl       dist/docker/ghcr
	cp -r docker/ghcr/grules     dist/docker/ghcr
	cp -r docker/ghcr/system     dist/docker/ghcr
	rsync -av --exclude='**/html.go' httpd/html dist/docker/ghcr
	cd dist/docker/ghcr && docker build --no-cache -f Dockerfile -t $(DOCKER) .

docker-compose: 
	rsync -av --exclude "html.go" httpd/html docker/compose/
	cd docker && find . -name .DS_Store -delete && rm -f compose.zip && zip --recurse-paths compose.zip compose

docker-run-dev:
	docker run --publish 8888:8080 --name httpd --rm uhppoted/uhppoted-httpd-dev
	sleep 1

docker-run-ghcr:
	docker run --publish 8888:8080 --publish 8443:8443 --name httpd --mount source=uhppoted-httpd,target=/usr/local/etc/uhppoted --rm ghcr.io/uhppoted/httpd
	sleep 1

docker-run-compose:
	cd docker/compose && docker compose up

docker-clean:
	docker image     prune -f
	docker container prune -f

docker-shell:
	docker exec -it httpd /bin/sh


