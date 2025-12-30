# PowerShell build script for URnetworkPC
# Run: .\build.ps1

$ErrorActionPreference = "Stop"

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "  URnetworkPC Build Script" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

# Check if Go is installed
try {
    $goVersion = go version
    Write-Host "Go version: $goVersion" -ForegroundColor Blue
    Write-Host ""
} catch {
    Write-Host "Error: Go is not installed" -ForegroundColor Red
    exit 1
}

# Download dependencies
Write-Host "Downloading dependencies..." -ForegroundColor Yellow
go mod download
go mod tidy
Write-Host "✓ Dependencies downloaded" -ForegroundColor Green
Write-Host ""

# Build development version (with console)
Write-Host "Building development version (with console)..." -ForegroundColor Yellow
go build -o URnetworkPC-dev.exe
Write-Host "✓ Development build complete: URnetworkPC-dev.exe" -ForegroundColor Green
Write-Host ""

# Build production version (without console, optimized)
Write-Host "Building production version (optimized)..." -ForegroundColor Yellow
go build -ldflags="-H windowsgui -s -w" -o URnetworkPC.exe
Write-Host "✓ Production build complete: URnetworkPC.exe" -ForegroundColor Green
Write-Host ""

# Show file sizes
Write-Host "Build artifacts:" -ForegroundColor Blue
Get-ChildItem URnetworkPC*.exe | Format-Table Name, @{Name="Size (MB)";Expression={[math]::Round($_.Length/1MB, 2)}}
Write-Host ""

Write-Host "=========================================" -ForegroundColor Green
Write-Host "  Build completed successfully!" -ForegroundColor Green
Write-Host "=========================================" -ForegroundColor Green
Write-Host ""
Write-Host "Files created:"
Write-Host "  • URnetworkPC-dev.exe  - Development version (with console for debugging)"
Write-Host "  • URnetworkPC.exe      - Production version (no console, optimized)"
Write-Host ""
Write-Host "To run: .\URnetworkPC.exe" -ForegroundColor Cyan
