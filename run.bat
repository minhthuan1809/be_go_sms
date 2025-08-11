@echo off
echo ⚡ Building and running SMS Gateway...
if not exist build mkdir build
go build -o build/sms-gateway.exe ./src/cmd/server
if %errorlevel% equ 0 (
    echo ✅ Build successful! Starting server...
    build/sms-gateway.exe
) else (
    echo ❌ Build failed!
)
