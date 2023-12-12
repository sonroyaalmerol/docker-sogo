#!/bin/bash
# Script to deploy on wiki docker-compose host
#

COMPOSE_CONTAINER="sogo"
COMPOSE_PROJECT="sogo"
COMPOSE_LOCATION="/srv/stacks/sogo/docker-compose.yml"

docker compose -f ${COMPOSE_LOCATION} -p ${COMPOSE_PROJECT} pull ${COMPOSE_CONTAINER}
docker compose -f ${COMPOSE_LOCATION} -p ${COMPOSE_PROJECT} up -d --force-recreate ${COMPOSE_CONTAINER} 