# Simple PowerShell test script
param(
  [string]$BaseUrl = "http://localhost:8080"
)

function Log($msg) { 
    Write-Host ("[" + (Get-Date).ToString("HH:mm:ss") + "] " + $msg) -ForegroundColor Cyan 
}

Log "Starting test script"
Log "Current time: $(Get-Date)"
Log "Base URL: $BaseUrl"

# Test basic HTTP request
Log "Testing HTTP connection..."
try {
    $response = Invoke-WebRequest -Uri "$BaseUrl/healthz" -Method GET -TimeoutSec 3
    Log "Connection successful: $($response.StatusCode)"
} catch {
    Log "Connection failed: $($_.Exception.Message)" -ForegroundColor Red
    Log "Please ensure server is running on $BaseUrl"
}

Log "Test completed" 