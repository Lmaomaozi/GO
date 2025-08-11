# PowerShell API 测试脚本
# 使用正确编码的修正版

param(
  [string]$BaseUrl = "http://localhost:8080",
  [string]$Phone = "13800138000"
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

function Format-Json {
    param($Object)
    $Object | ConvertTo-Json -Depth 8
}

# 检查服务器连通性
Write-Log "Checking server connectivity..."
try {
    $response = Invoke-WebRequest -Uri "$BaseUrl/healthz" -Method GET -TimeoutSec 5
    if ($response.StatusCode -eq 200) {
        Write-Log "Server connection successful"
    } else {
        Write-Log "Server response abnormal: $($response.StatusCode)"
    }
} catch {
    Write-Error-Log "Cannot connect to server $BaseUrl. Please ensure server is running."
    Write-Error-Log "Error details: $($_.Exception.Message)"
    exit 1
}

# 健康检查
Write-Log "Health check"
try {
    $health = Invoke-RestMethod -Uri "$BaseUrl/healthz" -Method GET
    Format-Json $health
} catch { 
    Write-Error-Log "Health check failed: $($_.Exception.Message)"
    exit 1
}

# 发送验证码
Write-Log "Sending verification code"
try {
    $body = @{ phone = $Phone } | ConvertTo-Json
    $sendRes = Invoke-RestMethod -Uri "$BaseUrl/api/user/send_code" -Method POST -ContentType "application/json" -Body $body
    $code = $sendRes.data.mock_code
    Format-Json $sendRes
} catch {
    Write-Error-Log "Failed to send verification code: $($_.Exception.Message)"
    exit 1
}

# 验证码登录
Write-Log "Login with verification code"
try {
    $body = @{ phone = $Phone; code = $code } | ConvertTo-Json
    $loginRes = Invoke-RestMethod -Uri "$BaseUrl/api/user/login" -Method POST -ContentType "application/json" -Body $body
    $access = $loginRes.data.accessToken
    $refresh = $loginRes.data.refreshToken
    Write-Log "accessToken (truncated): $($access.Substring(0,20))..."
    Format-Json $loginRes
} catch {
    Write-Error-Log "Login failed: $($_.Exception.Message)"
    exit 1
}

$headers = @{ Authorization = "Bearer $access" }

# 获取个人信息
Write-Log "Getting user info"
try {
    $userInfo = Invoke-RestMethod -Uri "$BaseUrl/api/user/me" -Headers $headers -Method GET
    Format-Json $userInfo
} catch {
    Write-Error-Log "Failed to get user info: $($_.Exception.Message)"
}

# 心跳上报
Write-Log "Sending heartbeat"
try {
    $heartbeat = Invoke-RestMethod -Uri "$BaseUrl/api/user/heartbeat" -Headers $headers -Method POST
    Format-Json $heartbeat
} catch {
    Write-Error-Log "Heartbeat failed: $($_.Exception.Message)"
}

# 一键登录
Write-Log "Testing one-click login"
try {
    $body = @{ phone = $Phone; device_id = "device-abc-123456"; platform = "android" } | ConvertTo-Json
    $oneclick = Invoke-RestMethod -Uri "$BaseUrl/api/user/oneclick_login" -Method POST -ContentType "application/json" -Body $body
    Format-Json $oneclick
} catch {
    Write-Error-Log "One-click login failed: $($_.Exception.Message)"
}

# 创建群组
Write-Log "Creating test group"
try {
    $body = @{ name = "Test Group"; avatar = "" } | ConvertTo-Json
    $groupRes = Invoke-RestMethod -Uri "$BaseUrl/api/group" -Headers $headers -Method POST -ContentType "application/json" -Body $body
    Format-Json $groupRes
} catch {
    Write-Error-Log "Failed to create group: $($_.Exception.Message)"
}

# 获取我的群组
Write-Log "Getting my groups"
try {
    $myGroups = Invoke-RestMethod -Uri "$BaseUrl/api/group/my" -Headers $headers -Method GET
    Format-Json $myGroups
} catch {
    Write-Error-Log "Failed to get groups: $($_.Exception.Message)"
}

# 上传头像
Write-Log "Testing avatar upload (requires images/avatar.jpg)"
$filePath = Join-Path (Get-Location) "images/avatar.jpg"
if (Test-Path $filePath) {
    try {
        # 使用正确的PowerShell文件上传语法
        $uploadRes = Invoke-WebRequest -Uri "$BaseUrl/api/file/avatar" -Headers $headers -Method POST -Form @{ file = Get-Item $filePath }
        Format-Json $uploadRes.Content
    } catch {
        Write-Error-Log "Avatar upload failed: $($_.Exception.Message)"
    }
} else {
    Write-Log "Skipping: File not found at $filePath"
}

# 关注状态
Write-Log "Checking follow status"
try {
    $me = Invoke-RestMethod -Uri "$BaseUrl/api/user/me" -Headers $headers -Method GET
    $myId = $me.data.user_id
    $followStatus = Invoke-RestMethod -Uri "$BaseUrl/api/relation/follow/status/$myId" -Headers $headers -Method GET
    Format-Json $followStatus
} catch {
    Write-Error-Log "Failed to get follow status: $($_.Exception.Message)"
}

# 消息操作
Write-Log "Testing message operations"
try {
    $convId = "dm_" + $myId + "_" + $myId
    $body = @{ conversation_type = "dm"; conversation_id = $convId; message_type = "user"; element = @{ type = "text"; text = "hello" } } | ConvertTo-Json
    $sendMsg = Invoke-RestMethod -Uri "$BaseUrl/api/message/send" -Headers $headers -Method POST -ContentType "application/json" -Body $body
    Format-Json $sendMsg

    $historyUrl = "$BaseUrl/api/message/history?conversation_type=dm&conversation_id=$convId&limit=10"
    $history = Invoke-RestMethod -Uri $historyUrl -Headers $headers -Method GET
    Format-Json $history
} catch {
    Write-Error-Log "Message operations failed: $($_.Exception.Message)"
}

# 刷新令牌
Write-Log "Testing token refresh"
try {
    $body = @{ refreshToken = $refresh } | ConvertTo-Json
    $refreshRes = Invoke-RestMethod -Uri "$BaseUrl/api/auth/refresh" -Method POST -ContentType "application/json" -Body $body
    Format-Json $refreshRes
} catch {
    Write-Error-Log "Token refresh failed: $($_.Exception.Message)"
}

Write-Log "All tests completed" 