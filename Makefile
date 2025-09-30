start:
	go build -o healthchecker.exe ./cmd/healthchecker/
	./healthchecker.exe -config-path=config.yaml