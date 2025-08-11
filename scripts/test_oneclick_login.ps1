# 一键登录测试脚本
param(
  [string]$BaseUrl = "http://localhost:8080"
)

function Write-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "HH:mm:ss"
    Write-Host "[$timestamp] $Message" -ForegroundColor Cyan
}

function Write-Error-Log {
    param([string]$Message)
    $timestamp = Get-Date -Format "HH:mm:ss"
    Write-Host "[$timestamp] ERROR: $Message" -ForegroundColor Red
}

Write-Log "Testing one-click login endpoint: $BaseUrl/api/user/oneclick_login"

# 请求载荷
$testData = @{
    phone = "13800138000"
    device_id = "test-device-123"
    platform = "android"
} | ConvertTo-Json

Write-Log "Request body: $testData"

try {
    # 使用 Invoke-RestMethod
    $response = Invoke-RestMethod -Uri "$BaseUrl/api/user/oneclick_login" -Method POST -ContentType "application/json" -Body $testData
    Write-Log "Request succeeded"
    Write-Log "Response: $($response | ConvertTo-Json -Depth 5)"
} catch {
    Write-Error-Log "Request failed: $($_.Exception.Message)"
    if ($_.Exception.Response) {
        Write-Error-Log "HTTP status: $($_.Exception.Response.StatusCode)"
        Write-Error-Log "Response content: $($_.Exception.Response.Content)"
    }
}

Write-Log "Test finished" 