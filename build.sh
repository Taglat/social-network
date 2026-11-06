#!/usr/bin/env bash
# Build and start both Docker images (backend + frontend).
set -e

cd "$(dirname "$0")"

echo "Building and starting the social network (backend + frontend)…"
docker compose up --build -d

echo
echo "Done. Containers:"
docker compose ps
echo
echo "Frontend: http://localhost:3000"
echo "Backend:  http://localhost:8080"
echo
echo "Stop with: docker compose down"
