#!/bin/bash
SCRIPT_DIR="$( cd -- "$( dirname -- "${BASH_SOURCE[0]}" )" &> /dev/null && pwd )"
cd $SCRIPT_DIR

help() {
  cat << EOF
This is a tool for development server managing.

Usage:
  ./manage.sh [command]

Available Commands:
  help        - this help
  start       - start server
  stop        - stop server
  purge       - stop and delete all containers

EOF
}

start() {
  echo ""
  docker-compose up -d

  sleep 1 && docker ps --format "table {{.ID}}\t{{.Names}}\t{{.Ports}}"
}

stop() {
  docker-compose stop
}

purge() {
  docker-compose stop
  docker-compose kill
  docker-compose rm -vf
}

if [[ $1 = "start" ]]; then
  start
  exit 0
fi

if [[ $1 = "stop" ]]; then
  stop
  exit 0
fi

if [[ $1 = "purge" ]]; then
  purge
  exit 0
fi

help