param(
    [Parameter(Mandatory = $true)]
    [string]$WorkspaceRoot,

    [string]$Profile = "",

    [string]$Flutter = "flutter"
)

$windowTitle = "Albz Dev Client"
if (-not [string]::IsNullOrWhiteSpace($Profile)) {
    $windowTitle = "Albz Dev Client $Profile"
}

$Host.UI.RawUI.WindowTitle = $windowTitle

Set-Location (Join-Path $WorkspaceRoot "app/frontend_flutter")

$arguments = @("run", "-d", "windows")
if (-not [string]::IsNullOrWhiteSpace($Profile)) {
    $arguments += "--dart-define=ALBZ_PROFILE=$Profile"
}

& $Flutter @arguments
