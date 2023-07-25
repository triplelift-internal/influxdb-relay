.PHONY: prune

build: build-dmp build-local

build-dmp:
	docker build -f Dockerfile.dmp \
--rm --tag "937685464166.dkr.ecr.eu-central-1.amazonaws.com/influxdb-relay:latest" .

build-local:
	docker build -f Dockerfile.local \
--rm --tag influxdb-relay:latest .

push-dmp:
	aws ecr get-login-password | \
docker login --username AWS --password-stdin 937685464166.dkr.ecr.eu-central-1.amazonaws.com
	docker push 937685464166.dkr.ecr.eu-central-1.amazonaws.com/influxdb-relay:latest

run:
	docker run -p 127.0.0.1:9096:9096 --rm influxdb-relay:latest

prune:
	docker image prune -a -f
