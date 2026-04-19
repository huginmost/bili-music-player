@echo off
setlocal

set ROOT=%~dp0
cd /d "%ROOT%"

if not exist release mkdir release
if not exist release\frontend mkdir release\frontend

echo Building bilig.exe...
go build -o bilig.exe ./cmd/bilig
if errorlevel 1 goto :fail

echo Building bmplayer-web.exe...
go build -o bmplayer-web.exe ./cmd/bmplayer-web
if errorlevel 1 goto :fail

echo Building frontend...
cd /d "%ROOT%frontend"
call npm run build
if errorlevel 1 goto :fail
cd /d "%ROOT%"

echo Copying files to release...
copy /Y bilig.exe release\bilig.exe >nul
if errorlevel 1 goto :copyfail
copy /Y bmplayer-web.exe release\bmplayer-web.exe >nul
if errorlevel 1 goto :copyfail
if exist bmpinfo.json (
  copy /Y bmpinfo.json release\bmpinfo.json >nul
  if errorlevel 1 goto :copyfail
)
xcopy /E /I /Y frontend\dist release\frontend >nul
if errorlevel 1 goto :copyfail

echo Release package updated.
exit /b 0

:copyfail
echo Packaging failed while copying files into release.
echo Please close any running executables in the release folder and try again.
exit /b 1

:fail
echo Packaging failed.
exit /b 1
