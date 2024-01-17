
param(
    [Parameter(Mandatory=$false)]
    [string]$args
)

$scriptVersion = "0.1.0"

switch ($args) {
    "" {
        $source = "github"
    }
    "github" {
        $source = "github"
    }
    "gitee" {
        $source = "gitee"
    }
    "--help" {
        $help = $true
    }
    "--version" {
        $version = $true
    }
    default {
        Write-Host "Unknown argument: $arg"
        Write-Host "Run 'install.ps1 --help' for usage."
        exit 1
    }
}

if ($help) {
    Write-Host "Usage: install.ps1 [options]"
    Write-Host "Options:"
    Write-Host "  --help             Show this help message and exit"
    Write-Host "  --version          Show script version info and exit"
    Write-Host "  gitee | github   Download from gitee/github, default: github"
    Write-Host ""
    Write-Host "Example:"
    Write-Host "  install.ps1 -source gitee"

    exit 0
}

if ($version) {
    Write-Host "OpenSCA install script version: $scriptVersion"
    exit 0
}

if ($source -eq "github") {
    $latest = Invoke-RestMethod -Uri "https://api.github.com/repos/XmirrorSecurity/OpenSCA-cli/releases/latest"

} else {
    $latest = Invoke-RestMethod -Uri "https://gitee.com/api/v5/repos/XmirrorSecurity/OpenSCA-cli/releases/latest"
}

# Get latesst version
$latestVersion = $latest.tag_name

# Get download URL
$downloadUrl = $latest.assets | Where-Object { $_.name -eq "opensca-cli-$latestVersion-windows-amd64.zip" } | Select-Object -ExpandProperty browser_download_url

# Get checksum URL
$checksumUrl = $latest.assets | Where-Object { $_.name -eq "opensca-cli-$latestVersion-windows-amd64.zip.sha256" } | Select-Object -ExpandProperty browser_download_url

# Print Download Info
Write-Host "* The latest version of OpenSCA-cli is: $latestVersion"

# Do Download
Write-Host "* Downloading OpenSCA-cli from: $downloadUrl"
Invoke-WebRequest -Uri $downloadUrl -OutFile "opensca-cli-$latestVersion.zip"
Invoke-WebRequest -Uri $checksumUrl -OutFile "opensca-cli-$latestVersion.zip.sha256"

# Verify Checksum
Write-Host "* Verifying checksum..."
$checksum = Get-FileHash -Path "opensca-cli-$latestVersion.zip" -Algorithm SHA256 | Select-Object -ExpandProperty Hash
$checksumExpected = Get-Content "opensca-cli-$latestVersion.zip.sha256"

if ($checksum -ne $checksumExpected) {
    Write-Error "  Checksum verification failed, please try again."
    Remove-Item -Path "opensca-cli-$latestVersion.zip" -Force
    Remove-Item -Path "opensca-cli-$latestVersion.zip.sha256" -Force
    exit 1
}

# Extract
Write-Host "* Extracting OpenSCA-cli..."
$openscaDir = Join-Path $env:USERPROFILE "opensca"
$attributes = [System.IO.FileAttributes]::Hidden
New-Item -ItemType Directory -Path $openscaDir -Force | Set-ItemProperty -Name Attributes -Value $attributes
Expand-Archive -Path "opensca-cli-$latestVersion.zip" -DestinationPath $openscaDir -Force

# Add to PATH
$systemPath = [Environment]::GetEnvironmentVariable("PATH", [EnvironmentVariableTarget]::User)

if (-not $systemPath.Contains($openscaDir)) {
    $newPath = $systemPath + ";" + $openscaDir
    Write-Host "* Adding OpenSCA-cli to PATH..."
    [Environment]::SetEnvironmentVariable("PATH", $newPath, [EnvironmentVariableTarget]::User)
}

Remove-Item -Path "opensca-cli-$latestVersion.zip" -Force
Remove-Item -Path "opensca-cli-$latestVersion.zip.sha256" -Force

Write-Host "* Successfully installed OpenSCA-cli in $extractPath. You can start using it by running 'opensca-cli' in your terminal. Enjoy!"
