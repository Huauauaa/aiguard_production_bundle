param(
  [string]$WorkspaceRoot = ""
)

$ErrorActionPreference = 'Stop'

if ([string]::IsNullOrWhiteSpace($WorkspaceRoot)) {
  $WorkspaceRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
}

$releaseDir = Join-Path $WorkspaceRoot 'release'
$binDir = Join-Path $WorkspaceRoot 'build\bin'
New-Item -ItemType Directory -Force -Path $releaseDir | Out-Null

if (-not (Test-Path $binDir)) {
  throw "未找到 build/bin，请先执行构建。"
}

$timestamp = Get-Date -Format 'yyyyMMdd-HHmmss'
$zipPath = Join-Path $releaseDir ("AIGuard-windows-$timestamp.zip")
if (Test-Path $zipPath) { Remove-Item $zipPath -Force }
Compress-Archive -Path (Join-Path $binDir '*') -DestinationPath $zipPath
Write-Host "已归档到 $zipPath" -ForegroundColor Green
