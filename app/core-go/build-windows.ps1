$ErrorActionPreference = "Stop"

function Resolve-GoExecutable {
    $candidates = @(
        (Join-Path $env:ProgramFiles "Go\bin\go.exe"),
        (Join-Path ${env:ProgramFiles(x86)} "Go\bin\go.exe")
    )

    foreach ($candidate in $candidates) {
        if ($candidate -and (Test-Path $candidate)) {
            return $candidate
        }
    }

    $command = Get-Command go.exe -ErrorAction SilentlyContinue
    if ($command -and $command.Source) {
        return $command.Source
    }

    throw "go.exe was not found"
}

function Resolve-GccExecutable {
    $candidates = @(
        "D:\MYSYS\ucrt64\bin\gcc.exe",
        "C:\msys64\ucrt64\bin\gcc.exe",
        "C:\msys64\mingw64\bin\gcc.exe"
    )

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    $command = Get-Command gcc.exe -ErrorAction SilentlyContinue
    if ($command -and $command.Source) {
        return $command.Source
    }

    throw "gcc.exe was not found"
}

function Resolve-GppExecutable {
    $candidates = @(
        "D:\MYSYS\ucrt64\bin\g++.exe",
        "C:\msys64\ucrt64\bin\g++.exe",
        "C:\msys64\mingw64\bin\g++.exe"
    )

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    $command = Get-Command g++.exe -ErrorAction SilentlyContinue
    if ($command -and $command.Source) {
        return $command.Source
    }

    throw "g++.exe was not found"
}

$repoRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$outputDir = Join-Path $repoRoot "app\frontend_flutter\native\windows"
$outputDLL = Join-Path $outputDir "albz_core.dll"

$goExe = Resolve-GoExecutable
$gccExe = Resolve-GccExecutable
$gppExe = Resolve-GppExecutable
$tempPath = [System.IO.Path]::GetTempPath()
$machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
$userPath = [Environment]::GetEnvironmentVariable("Path", "User")
$toolchainDir = Split-Path -Parent $gccExe
$goDir = Split-Path -Parent $goExe
$cleanPath = @($machinePath, $userPath, $goDir, $toolchainDir) -join ";"

New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

$startInfo = New-Object System.Diagnostics.ProcessStartInfo
$startInfo.FileName = $goExe
$startInfo.Arguments = "build -buildmode=c-shared -o `"$outputDLL`" ./ffi"
$startInfo.WorkingDirectory = $PSScriptRoot
$startInfo.UseShellExecute = $false
$startInfo.RedirectStandardOutput = $true
$startInfo.RedirectStandardError = $true

$startInfo.EnvironmentVariables.Clear()
$startInfo.EnvironmentVariables["PATH"] = $cleanPath
$startInfo.EnvironmentVariables["TEMP"] = $tempPath
$startInfo.EnvironmentVariables["TMP"] = $tempPath
$startInfo.EnvironmentVariables["SystemRoot"] = $env:SystemRoot
$startInfo.EnvironmentVariables["ComSpec"] = $env:ComSpec
$startInfo.EnvironmentVariables["PATHEXT"] = $env:PATHEXT
$startInfo.EnvironmentVariables["WINDIR"] = $env:WINDIR
$startInfo.EnvironmentVariables["USERPROFILE"] = $env:USERPROFILE
$startInfo.EnvironmentVariables["LOCALAPPDATA"] = $env:LOCALAPPDATA
$startInfo.EnvironmentVariables["APPDATA"] = $env:APPDATA
$startInfo.EnvironmentVariables["PROGRAMDATA"] = $env:PROGRAMDATA
$startInfo.EnvironmentVariables["CC"] = $gccExe
$startInfo.EnvironmentVariables["CXX"] = $gppExe

$process = New-Object System.Diagnostics.Process
$process.StartInfo = $startInfo

[void]$process.Start()
$stdout = $process.StandardOutput.ReadToEnd()
$stderr = $process.StandardError.ReadToEnd()
$process.WaitForExit()

if ($stdout) {
    Write-Output $stdout.TrimEnd()
}

if ($process.ExitCode -ne 0) {
    if ($stderr) {
        Write-Error $stderr.TrimEnd()
    }
    throw "go build failed with exit code $($process.ExitCode)"
}
