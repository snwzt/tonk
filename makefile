build:
	@go build -o bin/tt .

run:
	@./bin/tt

clean:
	@rm -rf bin