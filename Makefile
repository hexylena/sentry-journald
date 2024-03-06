sentry-journald:
	go build

fmt:
	go fmt

serve:
	find . -name '*.go' | entr -r go run main.go
