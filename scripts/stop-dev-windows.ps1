$markers = @(
    "scripts\run-server-window.ps1",
    "scripts\run-flutter-client-window.ps1"
)

$processes = Get-CimInstance Win32_Process | Where-Object {
    $commandLine = $_.CommandLine
    $_.Name -eq "powershell.exe" -and
    $commandLine -and
    (($markers | Where-Object { $commandLine -like ("*" + $_ + "*") }).Count -gt 0)
}

foreach ($process in $processes) {
    taskkill /PID $process.ProcessId /T /F | Out-Null
}
