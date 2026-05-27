$banner = @'
 █████╗ ██████╗  ██████╗██╗  ██╗ ██████╗ ███╗   ██╗
██╔══██╗██╔══██╗██╔════╝██║  ██║██╔═══██╗████╗  ██║
███████║██████╔╝██║     ███████║██║   ██║██╔██╗ ██║
██╔══██║██╔══██╗██║     ██╔══██║██║   ██║██║╚██╗██║
██║  ██║██║  ██║╚██████╗██║  ██║╚██████╔╝██║ ╚████║
╚═╝  ╚═╝╚═╝  ╚═╝ ╚═════╝╚═╝  ╚═╝ ╚═════╝ ╚═╝  ╚═══╝
'@

Write-Host $banner -ForegroundColor Cyan
Write-Host ""

$Owner = "sirrryasir"
$Repo = "archon"
$FallbackTag = "v0.1.1"

Write-Host "Checking the latest release of Archon..."
try {
    $ReleaseInfo = Invoke-RestMethod -Uri "https://api.github.com/repos/$Owner/$Repo/releases/latest" -UseBasicParsing
    $Tag = $ReleaseInfo.tag_name
} catch {
    $Tag = $FallbackTag
}

if ($null -eq $Tag -or $Tag -eq "") {
    $Tag = $FallbackTag
}

Write-Host "Latest release found: $Tag"

$Url = "https://github.com/sirrryasir/archon/releases/download/$Tag/archon-windows-amd64.exe"
$InstallDir = Join-Path $HOME ".archon\bin"
$DestPath = Join-Path $InstallDir "archon.exe"

if (-not (Test-Path $InstallDir)) {
    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
}

Write-Host "Downloading Archon for Windows..."
Invoke-WebRequest -Uri $Url -OutFile $DestPath -UseBasicParsing

$UserPath = [Environment]::GetEnvironmentVariable("Path", "User")
if ($UserPath -split ";" -notcontains $InstallDir) {
    Write-Host "Adding $InstallDir to your User PATH environment variable..."
    [Environment]::SetEnvironmentVariable("Path", $UserPath + ";" + $InstallDir, "User")
    $env:Path += ";$InstallDir"
}

Write-Host "Archon has been successfully installed to $DestPath!"
Write-Host "Please restart your terminal/PowerShell session and run 'archon doctor' to verify."
