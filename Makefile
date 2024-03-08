sentry-journald: $(wildcard *.go) $(wildcard templates/*)
	go build

fmt:
	go fmt

serve:
	find . -name '*.go' -or -name '*.html' | entr -r bash -c "go build && ./sentry-journald --port 8000"
