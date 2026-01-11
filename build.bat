@echo off
chcp 65001 >nul
setlocal enabledelayedexpansion

echo ========================================
echo   编译小红书图片生成系统
echo ========================================
echo.

REM 检查Go是否安装
where go >nul 2>&1
if errorlevel 1 (
    echo [错误] 未找到Go编译器，请确保Go已安装并添加到PATH环境变量中
    pause
    exit /b 1
)

REM 创建bin目录（如果不存在）
if not exist "bin" (
    echo 创建bin目录...
    mkdir bin
)

echo 正在下载依赖...
go mod download
if errorlevel 1 (
    echo [错误] 依赖下载失败
    pause
    exit /b 1
)

echo.
echo 正在编译API服务...
go build -o bin/api.exe ./cmd/api
if errorlevel 1 (
    echo [错误] API服务编译失败
    pause
    exit /b 1
)
echo [成功] API服务编译完成: bin/api.exe

echo.
echo 正在编译Worker服务...
go build -o bin/worker.exe ./cmd/worker
if errorlevel 1 (
    echo [错误] Worker服务编译失败
    pause
    exit /b 1
)
echo [成功] Worker服务编译完成: bin/worker.exe

echo.
echo ========================================
echo   编译完成！
echo ========================================
echo 可执行文件位置:
echo   - bin\api.exe
echo   - bin\worker.exe
echo.

pause