@echo off
setlocal enabledelayedexpansion

echo ================================
echo Docker Compose Build + Run
echo Using HTTP Proxy: 10809
echo ================================

REM ===== 代理配置 =====
set HTTP_PROXY=http://host.docker.internal:10809
set HTTPS_PROXY=http://host.docker.internal:10809
set NO_PROXY=localhost,127.0.0.1,*.local,host.docker.internal

REM ===== 显示当前代理（便于排错）=====
echo HTTP_PROXY=%HTTP_PROXY%
echo HTTPS_PROXY=%HTTPS_PROXY%
echo NO_PROXY=%NO_PROXY%
echo.

REM ===== 切到脚本所在目录（防止从别处执行）=====
cd /d %~dp0

REM ===== 构建镜像 =====
echo [1/2] docker compose build --no-cache
docker compose build --no-cache
if errorlevel 1 (
    echo.
    echo !!! Docker build failed !!!
    pause
    exit /b 1
)

REM ===== 启动服务 =====
echo.
echo [2/2] docker compose up -d
docker compose up -d
if errorlevel 1 (
    echo.
    echo !!! Docker compose up failed !!!
    pause
    exit /b 1
)

echo.
echo ================================
echo Docker Compose started SUCCESS
echo ================================
pause
