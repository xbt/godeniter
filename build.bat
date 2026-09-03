@echo off
REM ==============================================================================
REM Godeniter 2.0 Windows 原生一键构建脚本 (编译两个 Demo)
REM ==============================================================================

echo ==========================================================
echo  Building Godeniter 2.0 Demos for Windows
echo ==========================================================

if not exist dist mkdir dist

echo ^>^> 1. Compiling Demo 1 (API + SPA)...
go build -ldflags="-s -w" -o dist\demo1_api_spa.exe examples\01_api_spa\main.go

echo ^>^> 2. Compiling Demo 2 (MVC Template)...
go build -ldflags="-s -w" -o dist\demo2_mvc_template.exe examples\02_mvc_template\main.go

if %ERRORLEVEL% equ 0 (
    echo ==========================================================
    echo  Build Successful! Output in 'dist\' folder:
    echo  - dist\demo1_api_spa.exe
    echo  - dist\demo2_mvc_template.exe
    echo  You can double click either .exe to run directly.
    echo ==========================================================
) else (
    echo Build failed with error %ERRORLEVEL%
)

pause
