param(
    [Parameter(Mandatory = $true)]
    [string]$WorkspaceRoot
)

$Host.UI.RawUI.WindowTitle = "Haddle Dev Server"

Set-Location $WorkspaceRoot
$env:HADDLE_DEV_MODE = "1"

go run ./server
