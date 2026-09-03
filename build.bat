@echo off
REM ==============================================================================
REM Godeniter 2.0 Windows 原生一键构建脚本
REM ==============================================================================

echo ==========================================================
echo  Building Godeniter 2.0 for Windows
echo ==========================================================

if not exist dist mkdir dist

echo ^>^> Compiling single executable file...
go build -ldflags="-s -w" -o dist\godeniter-app.exe cmd\server\main.go

if %ERRORLEVEL% equ 0 (
    echo ==========================================================
    echo  Build Successful! Output: dist\godeniter-app.exe
    echo  You can now double click dist\godeniter-app.exe to run.
    echo ==========================================================
) else (
    echo Build failed with error %ERRORLEVEL%
)

pause
