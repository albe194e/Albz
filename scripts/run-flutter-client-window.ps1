param(
    [Parameter(Mandatory = $true)]
    [string]$WorkspaceRoot,

    [string]$Profile = ""
)

$exePath = Join-Path $WorkspaceRoot "app/frontend_flutter\build\windows\x64\runner\Debug\frontend_flutter.exe"

if (-not (Test-Path -LiteralPath $exePath)) {
    throw "Windows client executable not found at '$exePath'. Run 'make build-client-windows-debug' or 'make run-client' first."
}

$env:HADDLE_DEV_MODE = "1"
if ([string]::IsNullOrWhiteSpace($Profile)) {
    Remove-Item Env:\HADDLE_PROFILE -ErrorAction SilentlyContinue
} else {
    $env:HADDLE_PROFILE = $Profile
}

Start-Process -FilePath $exePath -WorkingDirectory (Split-Path -Parent $exePath)
