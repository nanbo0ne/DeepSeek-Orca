param(
  [string]$SourceDir = (Split-Path -Parent $MyInvocation.MyCommand.Path)
)

$ErrorActionPreference = "Stop"
$appName = "DeepCode"
$publisher = "DeepCode"
$version = "0.0.0"
$installDir = Join-Path $env:LOCALAPPDATA "Programs\DeepCode"
$exeName = "deepcode-desktop.exe"
$exePath = Join-Path $installDir $exeName
$uninstallPath = Join-Path $installDir "uninstall.exe"
$sourceExe = Join-Path $SourceDir $exeName
$sourceUninstall = Join-Path $SourceDir "uninstall.exe"
$sourceNode = Join-Path $SourceDir "node.exe"
$sourceCodegraph = Join-Path $SourceDir "codegraph"

Get-Process -Name "deepcode-desktop" -ErrorAction SilentlyContinue | Stop-Process -Force

New-Item -ItemType Directory -Force -Path $installDir | Out-Null
Copy-Item -LiteralPath $sourceExe -Destination $exePath -Force
Copy-Item -LiteralPath $sourceUninstall -Destination $uninstallPath -Force
if (Test-Path -LiteralPath $sourceNode) {
  Copy-Item -LiteralPath $sourceNode -Destination (Join-Path $installDir "node.exe") -Force
}
if (Test-Path -LiteralPath $sourceCodegraph) {
  $targetCodegraph = Join-Path $installDir "codegraph"
  Remove-Item -LiteralPath $targetCodegraph -Recurse -Force -ErrorAction SilentlyContinue
  Copy-Item -LiteralPath $sourceCodegraph -Destination $targetCodegraph -Recurse -Force
}

$shell = New-Object -ComObject WScript.Shell
$desktopLink = Join-Path ([Environment]::GetFolderPath("Desktop")) "$appName.lnk"
$startMenuDir = Join-Path ([Environment]::GetFolderPath("Programs")) $appName
New-Item -ItemType Directory -Force -Path $startMenuDir | Out-Null
$startMenuLink = Join-Path $startMenuDir "$appName.lnk"

foreach ($linkPath in @($desktopLink, $startMenuLink)) {
  Remove-Item -LiteralPath $linkPath -Force -ErrorAction SilentlyContinue
  $shortcut = $shell.CreateShortcut($linkPath)
  $shortcut.TargetPath = $exePath
  $shortcut.WorkingDirectory = $installDir
  $shortcut.IconLocation = "$exePath,0"
  $shortcut.Description = "DeepCode"
  $shortcut.Save()
}

$uninstKey = "HKCU:\Software\Microsoft\Windows\CurrentVersion\Uninstall\DeepCode"
New-Item -Force -Path $uninstKey | Out-Null
New-ItemProperty -Force -Path $uninstKey -Name "DisplayName" -Value $appName -PropertyType String | Out-Null
New-ItemProperty -Force -Path $uninstKey -Name "DisplayVersion" -Value $version -PropertyType String | Out-Null
New-ItemProperty -Force -Path $uninstKey -Name "Publisher" -Value $publisher -PropertyType String | Out-Null
New-ItemProperty -Force -Path $uninstKey -Name "InstallLocation" -Value $installDir -PropertyType String | Out-Null
New-ItemProperty -Force -Path $uninstKey -Name "DisplayIcon" -Value $exePath -PropertyType String | Out-Null
New-ItemProperty -Force -Path $uninstKey -Name "UninstallString" -Value "`"$uninstallPath`"" -PropertyType String | Out-Null
New-ItemProperty -Force -Path $uninstKey -Name "QuietUninstallString" -Value "`"$uninstallPath`" -Quiet" -PropertyType String | Out-Null

Start-Process -FilePath "$env:WINDIR\System32\ie4uinit.exe" -ArgumentList "-show" -WindowStyle Hidden -ErrorAction SilentlyContinue
Start-Process -FilePath $exePath
