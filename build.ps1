# Build and start both Docker images (backend + frontend).
$ErrorActionPreference = "Stop"
Set-Location $PSScriptRoot

Write-Host "Building and starting the social network (backend + frontend)..."
docker compose up --build -d

Write-Host ""
Write-Host "Done. Containers:"
docker compose ps
Write-Host ""
Write-Host "Frontend: http://localhost:3000"
Write-Host "Backend:  http://localhost:8080"
Write-Host ""
Write-Host "Stop with: docker compose down"
