param(
    [Parameter(Mandatory = $true)]
    [string]$WorkspaceRoot
)

$Host.UI.RawUI.WindowTitle = "Albz Dev Server"

Set-Location $WorkspaceRoot

go run ./server
