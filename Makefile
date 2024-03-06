serve:
	find . -name '*.go' | entr -r go run main.go
