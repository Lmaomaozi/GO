# PowerShell 测试脚本：按功能测试所有主要 API，并打印调试信息

param(
  [string]$BaseUrl = "http://localhost:8080",
  [string]$Phone = "13800138000"
)

function Log($msg) { Write-Host ("[" + (Get-Date).ToString("HH:mm:ss") + "] " + $msg) -ForegroundColor Cyan }
function Pretty($obj) { $obj | ConvertTo-Json -Depth 8 }

# 检查服务器是否可用
Log "检查服务器连接性..."
try {
  $response = Invoke-WebRequest -Uri "$BaseUrl/healthz" -Method GET -TimeoutSec 5
  if ($response.StatusCode -eq 200) {
    Log "服务器连接正常"
  } else {
    Log "服务器响应异常: $($response.StatusCode)"
  }
} catch {
  Log "无法连接到服务器 $BaseUrl，请确保服务器正在运行"
  Log "错误详情: $($_.Exception.Message)"
  exit 1
}

Log "健康检查"
try {
  $health = Invoke-RestMethod -Uri "$BaseUrl/healthz" -Method GET
  Pretty $health
} catch { 
  Log "健康检查失败: $($_.Exception.Message)" -ForegroundColor Red
  exit 1
}

Log "发送验证码"
try {
  $sendRes = Invoke-RestMethod -Uri "$BaseUrl/api/user/send_code" -Method POST -ContentType "application/json" -Body ( @{ phone = $Phone } | ConvertTo-Json )
  $code = $sendRes.data.mock_code
  Pretty $sendRes
} catch {
  Log "发送验证码失败: $($_.Exception.Message)" -ForegroundColor Red
  exit 1
}

Log "验证码登录"
try {
  $loginRes = Invoke-RestMethod -Uri "$BaseUrl/api/user/login" -Method POST -ContentType "application/json" -Body ( @{ phone = $Phone; code = $code } | ConvertTo-Json )
  $access = $loginRes.data.accessToken
  $refresh = $loginRes.data.refreshToken
  Log "accessToken(截断)：$($access.Substring(0,20))..."
  Pretty $loginRes
} catch {
  Log "登录失败: $($_.Exception.Message)" -ForegroundColor Red
  exit 1
}

$headers = @{ Authorization = "Bearer $access" }

Log "获取个人信息"
try {
  $userInfo = Invoke-RestMethod -Uri "$BaseUrl/api/user/me" -Headers $headers -Method GET
  Pretty $userInfo
} catch {
  Log "获取个人信息失败: $($_.Exception.Message)" -ForegroundColor Red
}

Log "心跳上报（在线状态）"
try {
  $heartbeat = Invoke-RestMethod -Uri "$BaseUrl/api/user/heartbeat" -Headers $headers -Method POST
  Pretty $heartbeat
} catch {
  Log "心跳上报失败: $($_.Exception.Message)" -ForegroundColor Red
}

Log "模拟一键登录（相同手机号与设备）"
try {
  $oneclick = Invoke-RestMethod -Uri "$BaseUrl/api/user/oneclick_login" -Method POST -ContentType "application/json" -Body ( @{ phone = $Phone; device_id = "device-abc-123456"; platform = "android" } | ConvertTo-Json )
  Pretty $oneclick
} catch {
  Log "一键登录失败: $($_.Exception.Message)" -ForegroundColor Red
}

Log "创建群组（示例）"
try {
  $groupRes = Invoke-RestMethod -Uri "$BaseUrl/api/group" -Headers $headers -Method POST -ContentType "application/json" -Body ( @{ name = "测试群组"; avatar = "" } | ConvertTo-Json )
  Pretty $groupRes
} catch {
  Log "创建群组失败: $($_.Exception.Message)" -ForegroundColor Red
}

Log "获取我的群组"
try {
  $myGroups = Invoke-RestMethod -Uri "$BaseUrl/api/group/my" -Headers $headers -Method GET
  Pretty $myGroups
} catch {
  Log "获取群组失败: $($_.Exception.Message)" -ForegroundColor Red
}

Log "上传头像（需要准备一张 images/avatar.jpg）"
$filePath = Join-Path (Get-Location) "images/avatar.jpg"
if (Test-Path $filePath) {
  try {
    $form = @{ file = Get-Item $filePath }
    $uploadRes = Invoke-RestMethod -Uri "$BaseUrl/api/file/avatar" -Headers $headers -Method POST -Form $form
    Pretty $uploadRes
  } catch {
    Log "上传头像失败: $($_.Exception.Message)" -ForegroundColor Red
  }
} else {
  Log "跳过：未找到 $filePath"
}

Log "关注状态（自测需准备另一个用户ID，示例使用自己ID会被禁止关注）"
try {
  $me = Invoke-RestMethod -Uri "$BaseUrl/api/user/me" -Headers $headers -Method GET
  $myId = $me.data.user_id
  $followStatus = Invoke-RestMethod -Uri "$BaseUrl/api/relation/follow/status/$myId" -Headers $headers -Method GET
  Pretty $followStatus
} catch {
  Log "获取关注状态失败: $($_.Exception.Message)" -ForegroundColor Red
}

Log "消息发送与拉取（示例会话ID使用 me_me）"
try {
  $convId = "dm_" + $myId + "_" + $myId
  $sendMsg = Invoke-RestMethod -Uri "$BaseUrl/api/message/send" -Headers $headers -Method POST -ContentType "application/json" -Body (@{ conversation_type = "dm"; conversation_id = $convId; message_type = "user"; element = @{ type = "text"; text = "hello" } } | ConvertTo-Json)
  Pretty $sendMsg

  # 修复URL中的&符号问题，使用变量存储URL
  $historyUrl = "$BaseUrl/api/message/history?conversation_type=dm&conversation_id=$convId&limit=10"
  $history = Invoke-RestMethod -Uri $historyUrl -Headers $headers -Method GET
  Pretty $history
} catch {
  Log "消息操作失败: $($_.Exception.Message)" -ForegroundColor Red
}

Log "手动刷新 Token"
try {
  $refreshRes = Invoke-RestMethod -Uri "$BaseUrl/api/auth/refresh" -Method POST -ContentType "application/json" -Body (@{ refreshToken = $refresh } | ConvertTo-Json)
  Pretty $refreshRes
} catch {
  Log "刷新Token失败: $($_.Exception.Message)" -ForegroundColor Red
}

Log "测试完成"

