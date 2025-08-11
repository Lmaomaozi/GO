# 头像上传测试脚本 - 兼容性版本
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
        Write-Log "文件路径: $filePath"
        Write-Log "文件大小: $((Get-Item $filePath).Length) bytes"

        # 使用 .NET HttpClient 构造 multipart/form-data，兼容 PowerShell 5.1
        Add-Type -AssemblyName System.Net.Http
        $client = New-Object System.Net.Http.HttpClient
        $client.DefaultRequestHeaders.Authorization = New-Object System.Net.Http.Headers.AuthenticationHeaderValue("Bearer", $access)

        $content = New-Object System.Net.Http.MultipartFormDataContent
        $fs = [System.IO.File]::OpenRead($filePath)
        try {
            $fileContent = New-Object System.Net.Http.StreamContent($fs)
            $fileContent.Headers.ContentType = New-Object System.Net.Http.Headers.MediaTypeHeaderValue("image/jpeg")
            $content.Add($fileContent, "file", "avatar.jpg")

            $resp = $client.PostAsync("$BaseUrl/api/file/avatar", $content).Result
            $respBody = $resp.Content.ReadAsStringAsync().Result

            Write-Log "上传HTTP状态: $([int]$resp.StatusCode)"
            Write-Host $respBody
        } finally {
            if ($fs) { $fs.Dispose() }
            if ($content) { $content.Dispose() }
            if ($client) { $client.Dispose() }
        }
    } catch {
        Write-Error-Log "头像上传失败: $($_.Exception.Message)"
        if ($_.Exception.Response) {
            Write-Error-Log "HTTP状态码: $($_.Exception.Response.StatusCode)"
        }
    }
} else {
    Write-Log "跳过：未找到头像文件 $filePath"
}

Write-Log "测试完成" 