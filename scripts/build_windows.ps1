param(
  [switch]$Nsis,
  [switch]$Sign,
  [string]$SignPfxPath = "",
  [string]$SignPfxPassword = "",
  [string]$TimestampUrl = "http://ts.ssl.com",
  [string]$WorkspaceRoot = ""
)

$ErrorActionPreference = 'Stop'

if ([string]::IsNullOrWhiteSpace($WorkspaceRoot)) {
  $WorkspaceRoot = (Resolve-Path (Join-Path $PSScriptRoot '..')).Path
}

Push-Location $WorkspaceRoot
try {
  ./scripts/preflight.ps1

  Write-Host "`n[1/5] 安装前端依赖" -ForegroundColor Cyan
  Push-Location frontend
  npm install
  npm run build
  Pop-Location

  Write-Host "`n[2/5] 生成 go.sum / 整理模块" -ForegroundColor Cyan
  go mod tidy

  Write-Host "`n[3/5] 运行 wails doctor" -ForegroundColor Cyan
  wails doctor

  Write-Host "`n[4/5] 构建 Windows EXE" -ForegroundColor Cyan
  wails build -platform windows/amd64

  if ($Nsis) {
    Write-Host "`n[4.5/5] 构建 NSIS 安装包" -ForegroundColor Cyan
    wails build -platform windows/amd64 -nsis
  }

  if ($Sign) {
    if ([string]::IsNullOrWhiteSpace($SignPfxPath) -or [string]::IsNullOrWhiteSpace($SignPfxPassword)) {
      throw "启用 -Sign 时必须提供 -SignPfxPath 和 -SignPfxPassword"
    }

    $signtool = Get-ChildItem 'C:\Program Files (x86)\Windows Kits\10\bin' -Recurse -Filter signtool.exe |
      Sort-Object FullName -Descending |
      Select-Object -First 1

    if ($null -eq $signtool) {
      throw "未找到 signtool.exe，请安装 Windows SDK"
    }

    Get-ChildItem ./build/bin -Filter *.exe | ForEach-Object {
      Write-Host "签名 $($_.FullName)" -ForegroundColor Yellow
      & $signtool.FullName sign /fd sha256 /tr $TimestampUrl /f $SignPfxPath /p $SignPfxPassword $_.FullName
    }
  }

  Write-Host "`n[5/5] 归档产物" -ForegroundColor Cyan
  ./scripts/package_release.ps1

  Write-Host "`n完成，产物位于 build/bin 和 release 目录。" -ForegroundColor Green
}
finally {
  Pop-Location
}
