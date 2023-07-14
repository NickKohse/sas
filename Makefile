dev:
	go run *.go
build:
	go build -o sas *.go
tests:
	./testing/test_runner.rb
run:
	./sas