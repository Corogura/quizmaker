#!/bin/bash

# Load environment variables from .env file
export $(cat .env | xargs)

docker run -p 8080:8080 \
  -e PORT="${PORT}" \
  -e DATABASE_URL="${DATABASE_URL}" \
  -e JWT_SECRET="${JWT_SECRET}" \
  corogura/quizmaker:latest
