param()

$ErrorActionPreference = 'Stop'

function Test-Command($name) {
  $cmd = Get-Command $name -ErrorAction SilentlyContinue
  if ($null -eq $cmd) {
    Write-Host "[FAIL] 缺少命令: $name" -ForegroundColor Red
    return $false
  }
  Write-Host "[ OK ] $name -> $($cmd.Source)" -ForegroundColor Green
  return $true
}

$all = $true
$all = (Test-Command go) -and $all
$all = (Test-Command git) -and $all
$all = (Test-Command npm) -and $all
$all = (Test-Command wails) -and $all

try {
  $wv2 = Get-ItemProperty 'HKLM:\SOFTWARE\Microsoft\EdgeUpdate\Clients\{F1E7E5D0-0B2B-4B15-9A0B-FA6FE7B9D28B}' -ErrorAction Stop
  Write-Host "[ OK ] 检测到 WebView2 Runtime" -ForegroundColor Green
} catch {
  Write-Host "[WARN] 未检测到 WebView2 Runtime，请先安装再构建" -ForegroundColor Yellow
}

if ($all) {
  Write-Host "`n运行 wails doctor..." -ForegroundColor Cyan
  wails doctor
} else {
  Write-Host "`n依赖不完整，请先补齐后重试。" -ForegroundColor Red
  exit 1
}
