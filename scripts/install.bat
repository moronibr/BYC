@echo off
REM Installation script for BYC on Windows

REM Set version
set VERSION=0.1.0

REM Create installation directory
set INSTALL_DIR=%USERPROFILE%\.byc
mkdir "%INSTALL_DIR%\bin" 2>nul

REM Download binaries (replace with actual download URLs)
echo Downloading BYC binaries...
REM Uncomment these lines when you have actual download URLs
REM powershell -Command "& {Invoke-WebRequest -Uri 'https://github.com/yourusername/byc/releases/download/v%VERSION%/bycnode_windows_amd64_%VERSION%.exe' -OutFile '%INSTALL_DIR%\bin\bycnode.exe'}"
REM powershell -Command "& {Invoke-WebRequest -Uri 'https://github.com/yourusername/byc/releases/download/v%VERSION%/bycminer_windows_amd64_%VERSION%.exe' -OutFile '%INSTALL_DIR%\bin\bycminer.exe'}"

REM Add to PATH
setx PATH "%PATH%;%INSTALL_DIR%\bin" /M

echo Installation complete!
echo You can now run BYC using:
echo   bycnode -type full -port 8333
echo   bycminer -node localhost:8333 -coin leah

pause 