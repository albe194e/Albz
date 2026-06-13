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

function Resolve-AndroidNdkRoot {
    $candidates = @(
        $env:ANDROID_NDK_HOME,
        $env:ANDROID_NDK_ROOT,
        (Join-Path $env:LOCALAPPDATA "Android\Sdk\ndk\28.2.13676358"),
        (Join-Path $env:LOCALAPPDATA "Android\Sdk\ndk\27.1.12297006")
    ) | Where-Object { $_ }

    foreach ($candidate in $candidates) {
        if (Test-Path $candidate) {
            return $candidate
        }
    }

    throw "Android NDK was not found"
}

function Invoke-AndroidBuild {
    param(
        [string]$GoExe,
        [string]$NdkRoot,
        [string]$GoArch,
        [string]$Abi,
        [string]$ClangName,
        [string]$OutputPath
    )

    $toolchainDir = Join-Path $NdkRoot "toolchains\llvm\prebuilt\windows-x86_64\bin"
    $clangPath = Join-Path $toolchainDir $ClangName
    if (!(Test-Path $clangPath)) {
        throw "Android clang toolchain was not found at $clangPath"
    }

    $tempPath = [System.IO.Path]::GetTempPath()
    $machinePath = [Environment]::GetEnvironmentVariable("Path", "Machine")
    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $goDir = Split-Path -Parent $GoExe
    $cleanPath = @($machinePath, $userPath, $goDir, $toolchainDir) -join ";"

    $outputDir = Split-Path -Parent $OutputPath
    New-Item -ItemType Directory -Force -Path $outputDir | Out-Null

    $startInfo = New-Object System.Diagnostics.ProcessStartInfo
    $startInfo.FileName = $GoExe
    $startInfo.Arguments = "build -buildmode=c-shared -o `"$OutputPath`" ./ffi"
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
    $startInfo.EnvironmentVariables["GOOS"] = "android"
    $startInfo.EnvironmentVariables["GOARCH"] = $GoArch
    $startInfo.EnvironmentVariables["CGO_ENABLED"] = "1"
    $startInfo.EnvironmentVariables["CC"] = $clangPath
    $startInfo.EnvironmentVariables["CXX"] = $clangPath

    $process = New-Object System.Diagnostics.Process
    $process.StartInfo = $startInfo

    Write-Output "Building Android ABI $Abi..."

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
        throw "go build failed for ABI $Abi with exit code $($process.ExitCode)"
    }
}

$repoRoot = Split-Path -Parent (Split-Path -Parent $PSScriptRoot)
$jniLibsRoot = Join-Path $repoRoot "app\frontend_flutter\android\app\src\main\jniLibs"
$goExe = Resolve-GoExecutable
$ndkRoot = Resolve-AndroidNdkRoot

Invoke-AndroidBuild `
    -GoExe $goExe `
    -NdkRoot $ndkRoot `
    -GoArch "arm64" `
    -Abi "arm64-v8a" `
    -ClangName "aarch64-linux-android21-clang.cmd" `
    -OutputPath (Join-Path $jniLibsRoot "arm64-v8a\libhaddle_core.so")

Invoke-AndroidBuild `
    -GoExe $goExe `
    -NdkRoot $ndkRoot `
    -GoArch "amd64" `
    -Abi "x86_64" `
    -ClangName "x86_64-linux-android21-clang.cmd" `
    -OutputPath (Join-Path $jniLibsRoot "x86_64\libhaddle_core.so")
