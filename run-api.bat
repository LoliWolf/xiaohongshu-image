@echo off
chcp 65001 >nul

echo ========================================
echo   启动API服务
echo ========================================
echo.

REM 切换到脚本所在目录
cd /d %~dp0

REM 检查.env文件是否存在
if not exist ".env" (
    echo [警告] 未找到 .env 文件
    if exist ".env.example" (
        echo [信息] 找到 .env.example 文件，正在复制...
        copy /Y .env.example .env >nul
        echo [信息] 已从 .env.example 创建 .env 文件
        echo [提示] 请编辑 .env 文件配置环境变量
        echo.
    ) else (
        echo [错误] 未找到 .env 或 .env.example 文件
        echo 请创建 .env 文件并配置环境变量
        echo.
        pause
        exit /b 1
    )
)

REM 加载.env文件中的环境变量
echo [信息] 正在加载环境变量...
if exist ".env" (
    for /f "usebackq eol=# tokens=1,* delims==" %%a in (".env") do (
        set "%%a=%%b"
    )
    echo [信息] 环境变量加载完成
) else (
    echo [警告] 无法加载环境变量文件
)

REM 检查可执行文件是否存在
if not exist "bin\api.exe" (
    echo [错误] 未找到可执行文件 bin\api.exe
    echo 请先运行 build.bat 编译项目
    echo.
    pause
    exit /b 1
)

REM 检查配置文件是否存在
if not exist "config\config.yaml" (
    echo [错误] 未找到配置文件 config\config.yaml
    echo 请确保配置文件存在
    echo.
    pause
    exit /b 1
)

echo [信息] 配置文件: config\config.yaml
echo [信息] 正在启动API服务...
echo [提示] 如需调试，请使用VSCode/Cursor的调试功能（F5）
echo.

REM 运行API服务
bin\api.exe

REM 如果程序退出，显示退出信息
if errorlevel 1 (
    echo.
    echo [错误] API服务异常退出 (错误代码: %errorlevel%)
    pause
    exit /b %errorlevel%
)
