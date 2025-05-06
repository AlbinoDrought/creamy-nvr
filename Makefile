all: ui creamy-nvr

.PHONY: ui
ui: 
	cd ui && npm ci && npm run build

.PHONY: creamy-nvr
creamy-nvr:
	# go test ./...
	git archive HEAD -o ui/dist/source.tar.gz
	go build -o creamy-nvr
