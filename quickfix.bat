@echo off
echo 🔧 Quick Fix for Veloci Mail
echo ========================

echo Step 1: Cleaning up...
if exist cmd\root.go (
    echo - Removing empty cmd\root.go
    del cmd\root.go
)
if exist cmd rmdir cmd 2>nul

echo Step 2: Fixing dependencies...
go clean -modcache
go mod tidy

echo Step 3: Testing compilation...
go build .
if %ERRORLEVEL% neq 0 (
    echo ❌ Compilation failed. Check the errors above.
    pause
    exit /b 1
)

echo ✅ Quick fix complete!
echo.
echo Next steps:
echo 1. Run: go run .
echo 2. Follow OAuth setup instructions
echo 3. Enjoy your email client!
echo.
pause
