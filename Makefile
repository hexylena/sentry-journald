sentry-journald:
	go build

fmt:
	go fmt

serve:
	find . -name '*.go' | entr -r bash -c "go build && ./sentry-journald --port 8000"
