param(
    [Parameter(Mandatory = $true)]
    [string]$WorkspaceRoot,

    [string]$Profile = "",

    [string]$Flutter = "flutter"
)

$windowTitle = "Haddle Dev Client"
if (-not [string]::IsNullOrWhiteSpace($Profile)) {
    $windowTitle = "Haddle Dev Client $Profile"
}

$Host.UI.RawUI.WindowTitle = $windowTitle

Set-Location (Join-Path $WorkspaceRoot "app/frontend_flutter")

$arguments = @("run", "-d", "windows", "--dart-define=HADDLE_DEV_MODE=true")
if (-not [string]::IsNullOrWhiteSpace($Profile)) {
    $arguments += "--dart-define=HADDLE_PROFILE=$Profile"
}

& $Flutter @arguments
