.PHONY: all build clean prune

all: build push-dmp

dmp: build-dmp push-dmp

local: build-local run

build: build-local build-dmp

build-local:
	docker build -f Dockerfile.local \
--rm --tag influxdb-relay:latest .

build-dmp:
	docker build -f Dockerfile.dmp \
--rm --tag "937685464166.dkr.ecr.eu-central-1.amazonaws.com/influxdb-relay:latest" .

push-dmp:
	aws ecr get-login-password | \
docker login --username AWS --password-stdin 937685464166.dkr.ecr.eu-central-1.amazonaws.com
	docker push 937685464166.dkr.ecr.eu-central-1.amazonaws.com/influxdb-relay:latest

# eg. make redeploy-dmp cluster=dev region=eu-central-1
redeploy-dmp:
	aws ecs update-service --cluster $(cluster) --service influxdb-relay --region $(region) --force-new-deployment

run: build-local
	docker run \
-p 127.0.0.1:9096:9096 \
-p 127.0.0.1:36936:36936/udp \
--rm influxdb-relay:latest

clean: prune
prune:
	docker image prune -a -f

