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
if (Test-Path $profilePath) {
    $content = Get-Content $profilePath | Where-Object { $_ -notmatch [regex]::Escape($importLine) }
    Set-Content -Path $profilePath -Value $content
}