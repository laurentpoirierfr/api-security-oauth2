clean-docker:
	./scripts/clean-docker.sh

start-docker:
	cd infras && docker-compose up