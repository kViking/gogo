param(
    [string]$GadgetPath
)

$profilePath = $PROFILE
if (!(Test-Path $profilePath)) {
    New-Item -ItemType File -Path $profilePath -Force | Out-Null
    Write-Host "Created new PowerShell profile at $profilePath"
}

if (-not $GadgetPath) {
    Write-Host "No GoGoGadget.ps1 path provided. Exiting."
    exit 1
}

$importLine = ". \"$GadgetPath\""

# Check if the import line already exists
if (-not (Select-String -Path $profilePath -Pattern ([regex]::Escape($importLine)) -Quiet)) {
    # Add the import line to the profile
    Add-Content -Path $profilePath -Value $importLine
    Write-Host "Added import line to PowerShell profile"
} else {
    Write-Host "Import line already exists in profile"
}

# Verify the content was written
if (Select-String -Path $profilePath -Pattern ([regex]::Escape($importLine)) -Quiet) {
    Write-Host "Verified: Import line is in the profile"
} else {
    Write-Host "Warning: Import line was not found in profile after adding it"
}