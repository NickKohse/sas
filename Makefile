dev:
	go run *.go
build:
	go build -o sas *.go
test:
	./testing/test_runner.rb
run:
	./sas