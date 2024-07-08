build:
	docker build --tag ghcr.io/some0ne3/ziu-checker:latest .
build-all:
	docker buildx build --platform linux/amd64,linux/arm64 --tag ghcr.io/some0ne3/ziu-checker:latest --push .

publish:
	docker push ghcr.io/some0ne3/ziu-checker:latest
