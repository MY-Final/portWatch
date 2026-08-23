# PortWatch installer for Windows.
#
# Usage:
#   .\install.ps1                      install the latest release
#   .\install.ps1 -Version v0.7.0      install a pinned release tag
#   .\install.ps1 -Uninstall           remove portwatch from %USERPROFILE%\bin
#
# Downloaded from GitHub releases of https://github.com/MY-Final/portWatch.
#requires -Version 5.1
[CmdletBinding()]
param(
    [string]$Version = "latest",
    [switch]$Uninstall
)

$ErrorActionPreference = 'Stop'
$ProgressPreference = 'SilentlyContinue'

$RepoUrl = 'https://github.com/MY-Final/portWatch'
$ApiBase = 'https://api.github.com/repos/MY-Final/portWatch'
$InstallDir = Join-Path $env:USERPROFILE 'bin'
$Target = Join-Path $InstallDir 'portwatch.exe'

function Write-Step([string]$Message) { Write-Host "==> $Message" }

# Broadcast WM_SETTINGCHANGE so running shells notice PATH updates without
# logging out again. Best-effort: failures are ignored.
function Broadcast-EnvironmentChange {
    try {
        if (-not ('PortWatch.EnvRefresh' -as [type])) {
            Add-Type -Namespace PortWatch -Name EnvRefresh -MemberDefinition @'
[DllImport("user32.dll", SetLastError = true, CharSet = CharSet.Auto)]
public static extern IntPtr SendMessageTimeout(IntPtr hWnd, uint Msg, UIntPtr wParam, string lParam, uint fuFlags, uint uTimeout, out UIntPtr lpdwResult);
'@
        }
        $result = [UIntPtr]::Zero
        [PortWatch.EnvRefresh]::SendMessageTimeout(
            [IntPtr]0xFFFF, 0x1A, [UIntPtr]::Zero, 'Environment', 2, 5000, [ref]$result) | Out-Null
    } catch { }
}

# Read and write the user PATH through the registry so the raw value keeps
# entries such as %JAVA_HOME%\bin unexpanded. [Environment]::GetEnvironment
# Variable would expand them and a write-back would destroy those entries.
function Open-UserEnvironmentKey([bool]$Writable) {
    if ($Writable) {
        return [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment', $true)
    }
    return [Microsoft.Win32.Registry]::CurrentUser.OpenSubKey('Environment')
}

function Get-RawUserPath {
    $key = Open-UserEnvironmentKey $false
    if ($null -eq $key) { return $null }
    try {
        return $key.GetValue('Path', $null,
            [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
    } finally { $key.Close() }
}

function Add-UserPathEntry([string]$Dir) {
    $key = Open-UserEnvironmentKey $true
    if ($null -eq $key) {
        Write-Warning "cannot open HKCU\Environment; add $Dir to PATH manually"
        return
    }
    try {
        $kind = [Microsoft.Win32.RegistryValueKind]::ExpandString
        try { $kind = $key.GetValueKind('Path') } catch { }
        $raw = $key.GetValue('Path', $null,
            [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
        if ($null -eq $raw -or ($raw -split ';').Trim() -notcontains $Dir) {
            $updated = if ([string]::IsNullOrWhiteSpace($raw)) { $Dir } else { $raw.TrimEnd(';') + ';' + $Dir }
            $key.SetValue('Path', $updated, $kind)
            Broadcast-EnvironmentChange
            Write-Host "Added $Dir to the user PATH."
        }
    } finally { $key.Close() }
}

function Remove-UserPathEntry([string]$Dir) {
    $key = Open-UserEnvironmentKey $true
    if ($null -eq $key) {
        Write-Warning "cannot open HKCU\Environment; remove $Dir from PATH manually"
        return
    }
    try {
        $kind = [Microsoft.Win32.RegistryValueKind]::ExpandString
        try { $kind = $key.GetValueKind('Path') } catch { }
        $raw = $key.GetValue('Path', $null,
            [Microsoft.Win32.RegistryValueOptions]::DoNotExpandEnvironmentNames)
        if ($null -eq $raw) { return }
        $dirNorm = $Dir.TrimEnd('\')
        $entries = @($raw -split ';')
        $kept = @($entries | Where-Object { $_.Trim().TrimEnd('\') -ne $dirNorm })
        if ($kept.Count -eq $entries.Count) { return }
        $updated = ($kept -join ';').Trim(';')
        if ([string]::IsNullOrWhiteSpace($updated)) {
            $key.DeleteValue('Path', $false)
        } else {
            $key.SetValue('Path', $updated, $kind)
        }
        Broadcast-EnvironmentChange
        Write-Host "Removed $Dir from the user PATH."
    } finally { $key.Close() }
}

if ($Uninstall) {
    if (-not (Test-Path -LiteralPath $Target)) {
        Write-Host "portwatch not found at $Target (already uninstalled)."
        exit 0
    }
    Write-Step "Deleting $Target"
    try {
        Remove-Item -LiteralPath $Target -Force
    } catch {
        Write-Error "failed to delete ${Target}: $($_.Exception.Message); close any running portwatch process and retry"
        exit 1
    }
    $remaining = @(Get-ChildItem -LiteralPath $InstallDir -Force -ErrorAction SilentlyContinue)
    if ($remaining.Count -eq 0) {
        Remove-UserPathEntry $InstallDir
        # No -Recurse: Remove-Item on a non-empty directory fails, so a file
        # that appears between the count and this call keeps the directory.
        Remove-Item -LiteralPath $InstallDir -ErrorAction SilentlyContinue
        if (-not (Test-Path -LiteralPath $InstallDir)) {
            Write-Host "Removed empty directory $InstallDir."
        }
    }
    Write-Host "portwatch uninstalled."
    exit 0
}

Write-Step "Resolving release ($Version)"
try {
    if ($Version -eq 'latest') {
        $release = Invoke-RestMethod -Uri "$ApiBase/releases/latest"
    } else {
        $tag = if ($Version.StartsWith('v')) { $Version } else { "v$Version" }
        $release = Invoke-RestMethod -Uri "$ApiBase/releases/tags/$tag"
    }
} catch {
    Write-Error "could not resolve the release from $ApiBase ($Version): $($_.Exception.Message)"
    exit 1
}

$arch = if ($env:PROCESSOR_ARCHITECTURE -eq 'ARM64') { 'arm64' } else { 'amd64' }
$wanted = "windows_${arch}.zip"
$asset = @($release.assets) | Where-Object { $_.name.EndsWith($wanted) } | Select-Object -First 1
if (-not $asset) {
    Write-Error "release $($release.tag_name) has no $wanted asset; check $RepoUrl/releases"
    exit 1
}
$checksumAsset = @($release.assets) | Where-Object { $_.name -eq 'checksums.txt' } | Select-Object -First 1

$tempDir = Join-Path ([System.IO.Path]::GetTempPath()) ("portwatch-install-" + [guid]::NewGuid().ToString('N'))
try {
    New-Item -ItemType Directory -Path $tempDir | Out-Null
    Write-Step "Downloading $($asset.name)"
    $zipPath = Join-Path $tempDir $asset.name
    try {
        Invoke-WebRequest -Uri $asset.browser_download_url -OutFile $zipPath
    } catch {
        Write-Error "download failed for $($asset.browser_download_url): $($_.Exception.Message)"
        exit 1
    }

    if ($checksumAsset) {
        $checksumPath = Join-Path $tempDir 'checksums.txt'
        try {
            Invoke-WebRequest -Uri $checksumAsset.browser_download_url -OutFile $checksumPath
        } catch {
            Write-Error "download failed for checksums.txt: $($_.Exception.Message)"
            exit 1
        }
        $expected = Get-Content -LiteralPath $checksumPath |
            Where-Object { $_ -match "\s$([regex]::Escape($asset.name))$" } |
            ForEach-Object { ($_ -split '\s+')[0] } | Select-Object -First 1
        if (-not $expected) {
            Write-Error "checksums.txt has no entry for $($asset.name)"
            exit 1
        }
        $actual = (Get-FileHash -LiteralPath $zipPath -Algorithm SHA256).Hash.ToLowerInvariant()
        if ($actual -ne $expected.ToLowerInvariant()) {
            Write-Error "SHA256 mismatch for $($asset.name): expected $expected, got $actual"
            exit 1
        }
        Write-Host "Checksum verified."
    } else {
        Write-Warning "release $($release.tag_name) has no checksums.txt; skipping verification"
    }

    $extractDir = Join-Path $tempDir 'extract'
    Expand-Archive -LiteralPath $zipPath -DestinationPath $extractDir
    $exe = Get-ChildItem -LiteralPath $extractDir -Recurse -Filter 'portwatch.exe' | Select-Object -First 1
    if (-not $exe) {
        Write-Error "archive $($asset.name) contains no portwatch.exe"
        exit 1
    }

    New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null
    try {
        Copy-Item -LiteralPath $exe.FullName -Destination $Target -Force
    } catch {
        Write-Error "cannot replace $Target (is portwatch running?): $($_.Exception.Message)"
        exit 1
    }
} finally {
    Remove-Item -LiteralPath $tempDir -Recurse -Force -ErrorAction SilentlyContinue
}

Add-UserPathEntry $InstallDir

Write-Step "Installed"
try {
    & $Target --version
} catch {
    Write-Warning "installed, but could not print the version: $($_.Exception.Message)"
}
Write-Host "Installed to $Target"
