# 头像上传测试脚本
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

# 1. 发送验证码
Write-Log "发送验证码"
try {
    $body = @{ phone = $Phone } | ConvertTo-Json
    $sendRes = Invoke-RestMethod -Uri "$BaseUrl/api/user/send_code" -Method POST -ContentType "application/json" -Body $body
    $code = $sendRes.data.mock_code
    Write-Log "验证码: $code"
} catch {
    Write-Error-Log "发送验证码失败: $($_.Exception.Message)"
    exit 1
}

# 2. 登录获取token
Write-Log "登录获取token"
try {
    $body = @{ phone = $Phone; code = $code } | ConvertTo-Json
    $loginRes = Invoke-RestMethod -Uri "$BaseUrl/api/user/login" -Method POST -ContentType "application/json" -Body $body
    $access = $loginRes.data.accessToken
    Write-Log "登录成功，token: $($access.Substring(0,20))..."
} catch {
    Write-Error-Log "登录失败: $($_.Exception.Message)"
    exit 1
}

$headers = @{ Authorization = "Bearer $access" }

# 3. 测试头像上传
Write-Log "测试头像上传"
$filePath = Join-Path (Get-Location) "images/avatar.jpg"
if (Test-Path $filePath) {
    try {
        # 使用正确的PowerShell文件上传语法
        $file = Get-Item $filePath
        $form = @{ file = $file }
        
        Write-Log "文件路径: $filePath"
        Write-Log "文件大小: $($file.Length) bytes"
        
        # 使用Invoke-WebRequest进行文件上传
        $uploadRes = Invoke-WebRequest -Uri "$BaseUrl/api/file/avatar" -Headers $headers -Method POST -Form $form
        
        Write-Log "上传成功，状态码: $($uploadRes.StatusCode)"
        Write-Log "响应内容: $($uploadRes.Content)"
        
    } catch {
        Write-Error-Log "头像上传失败: $($_.Exception.Message)"
        Write-Error-Log "错误详情: $($_.Exception.Response.StatusCode)"
    }
} else {
    Write-Log "跳过：未找到头像文件 $filePath"
}

Write-Log "测试完成" 