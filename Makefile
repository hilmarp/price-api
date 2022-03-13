.PHONY: build-scraper run-scraper
.PHONY: docker-up

build-scraper:
	go build -o bin/scraper cmd/scraper/main.go

run-scraper: build-scraper
	./bin/scraper

docker-up:
	docker-compose up