.PHONY: frontend embed test test-backend test-frontend docker-up

frontend:
	cd web && npm install && npm run build

embed: frontend
	rm -rf cmd/server/static
	mkdir -p cmd/server/static
	cp -R web/dist/. cmd/server/static/

test-backend:
	go test ./...

test-frontend:
	cd web && npm install && npm run test

test: test-backend test-frontend

docker-up: embed
	docker compose up -d --build
