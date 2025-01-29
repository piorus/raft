run-docker:
	docker compose build && docker compose up
run-docker-single:
	docker compose build && docker compose up server-1
run-local:
	go run main.go --port 1234