@echo off
setlocal EnableExtensions

set "APP_NAME=DeepCode"
set "EXE_NAME=deepcode-desktop.exe"
set "APPDIR=%~dp0"
for %%I in ("%APPDIR%.") do set "APPDIR=%%~fI"

set "CLEANER=%TEMP%\DeepCode-uninstall-%RANDOM%-%RANDOM%.cmd"

> "%CLEANER%" (
  echo @echo off
  echo setlocal EnableExtensions
  echo taskkill /IM "%EXE_NAME%" /F ^>nul 2^>nul
  echo powershell -NoProfile -ExecutionPolicy Bypass -Command "Get-Process node -ErrorAction SilentlyContinue ^| Where-Object { $_.Path -like '%APPDIR%*' } ^| Stop-Process -Force" ^>nul 2^>nul
  echo timeout /t 1 /nobreak ^>nul 2^>nul
  echo del /f /q "%USERPROFILE%\Desktop\%APP_NAME%.lnk" ^>nul 2^>nul
  echo del /f /q "%APPDATA%\Microsoft\Windows\Start Menu\Programs\%APP_NAME%.lnk" ^>nul 2^>nul
  echo del /f /q "%APPDATA%\Microsoft\Windows\Start Menu\Programs\Uninstall %APP_NAME%.lnk" ^>nul 2^>nul
  echo rmdir /s /q "%APPDATA%\%EXE_NAME%" ^>nul 2^>nul
  echo reg delete "HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\%APP_NAME%" /f ^>nul 2^>nul
  echo reg delete "HKCU\Software\Microsoft\Windows\CurrentVersion\Uninstall\%EXE_NAME%" /f ^>nul 2^>nul
  echo cd /d "%TEMP%" ^>nul 2^>nul
  echo rmdir /s /q "%APPDIR%" ^>nul 2^>nul
  echo del /f /q "%%~f0" ^>nul 2^>nul
)

start "" /min cmd.exe /c "%CLEANER%"
exit /b 0
