build:
	@go build -o bin/tonk .

run:
	@./bin/tonk

clean:
	@rm -rf bin