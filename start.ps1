Write-Host "Starting frontend and backend..."
Push-Location (Join-Path $PSScriptRoot "backend")
go run .
Pop-Location
