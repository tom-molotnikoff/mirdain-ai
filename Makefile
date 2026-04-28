.PHONY: build ui-build run test image e2e

BIN := bin/mirdain

build: ui-build
	mkdir -p bin
	CGO_ENABLED=0 go build -o $(BIN) ./cmd/mirdain

ui-build:
	cd ui && npm install && npm run build

run: build
	./$(BIN)

test:
	go test ./...

image:
	docker build -t mirdain-base .

e2e:
	@echo "e2e: stub — no tests yet"
