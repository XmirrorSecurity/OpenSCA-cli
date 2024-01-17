
# Get Latest Release Info
$latest = Invoke-RestMethod -Uri https://api.github.com/repos/XmirrorSecurity/OpenSCA-cli/releases/latest

# Latest Release Version
$latestVersion = $latest.tag_name

# Package Download URL
$downloadUrl = $latest.assets | Where-Object { $_.name -eq "opensca-cli-$latestVersion-windows-amd64.zip" } | Select-Object -ExpandProperty browser_download_url

# Checksum Download URL
$checksumUrl = $latest.assets | Where-Object { $_.name -eq "opensca-cli-$latestVersion-windows-amd64.zip.sha256" } | Select-Object -ExpandProperty browser_download_url

# Print Download Info
Write-Host "Latest Version: $latestVersion"
Write-Host "Download URL: $downloadUrl"
Write-Host "Checksum URL: $checksumUrl"

# Do Download
Invoke-WebRequest -Uri $downloadUrl -OutFile "opensca-cli-$latestVersion.zip"
Invoke-WebRequest -Uri $checksumUrl -OutFile "opensca-cli-$latestVersion.zip.sha256"

# Verify Checksum
$checksum = Get-FileHash -Path "opensca-cli-$latestVersion.zip" -Algorithm SHA256 | Select-Object -ExpandProperty Hash
$checksumExpected = Get-Content "opensca-cli-$latestVersion.zip.sha256"

if ($checksum -ne $checksumExpected) {
    Write-Error "Checksum verification failed"
    exit 1
}

# Extract
$openscaDir = Join-Path $env:USERPROFILE "opensca"
$attributes = [System.IO.FileAttributes]::Hidden
New-Item -ItemType Directory -Path $openscaDir -Force | Set-ItemProperty -Name Attributes -Value $attributes
Expand-Archive -Path "opensca-cli-$latestVersion.zip" -DestinationPath $openscaDir

# Add to PATH
$systemPath = [Environment]::GetEnvironmentVariable("PATH", [EnvironmentVariableTarget]::User)
if (-not $systemPath.Contains($extractPath)) {
    $newPath = $systemPath + ";" + $extractPath
    [Environment]::SetEnvironmentVariable("PATH", $newPath, [EnvironmentVariableTarget]::User)
}

# 输出结果
Write-Host "OpenSCA CLI installed and added to user PATH successfully."
