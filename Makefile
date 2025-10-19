all: ui creamy-nvr

.PHONY: ui
ui: 
	cd ui && npm ci && cp node_modules/@ffmpeg/core/dist/esm/* public/ffmpeg/ && npm run build

.PHONY: creamy-nvr
creamy-nvr:
	# go test ./...
	git archive HEAD -o ui/dist/source.tar.gz
	go build -o creamy-nvr
