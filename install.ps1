#Requires -Version 5.1

$ErrorActionPreference = "Stop"

$Repo = if ($env:TOTP_REPO) { $env:TOTP_REPO } else { "robby031/totp-cli" }

$rawArch = if ($env:PROCESSOR_ARCHITEW6432) { $env:PROCESSOR_ARCHITEW6432 } else { $env:PROCESSOR_ARCHITECTURE }
$arch = switch ($rawArch.ToUpperInvariant()) {
    "AMD64" { "amd64" }
    "ARM64" { "arm64" }
    default { throw "Arsitektur $rawArch tidak didukung" }
}

if ($env:TOTP_VERSION) {
    $baseUrl = "https://github.com/$Repo/releases/download/$($env:TOTP_VERSION)"
} else {
    $baseUrl = "https://github.com/$Repo/releases/latest/download"
}
if ($env:TOTP_BASE_URL) {
    $baseUrl = $env:TOTP_BASE_URL
}

$archive = "totp-windows-$arch.zip"
$tmp = Join-Path ([IO.Path]::GetTempPath()) ("totp-cli-" + [Guid]::NewGuid().ToString("N"))
New-Item -ItemType Directory -Path $tmp -Force | Out-Null

try {
    Write-Host "Mengunduh totp-cli windows/$arch..."
    Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/$archive" -OutFile (Join-Path $tmp $archive)
    Invoke-WebRequest -UseBasicParsing -Uri "$baseUrl/checksums.txt" -OutFile (Join-Path $tmp "checksums.txt")

    $line = Get-Content (Join-Path $tmp "checksums.txt") |
        Where-Object { ($_ -split "\s+")[1] -eq $archive } |
        Select-Object -First 1
    if (-not $line) {
        throw "Checksum untuk $archive tidak ditemukan"
    }
    $expected = ($line -split "\s+")[0]
    $actual = (Get-FileHash -Algorithm SHA256 -Path (Join-Path $tmp $archive)).Hash.ToLowerInvariant()
    if ($actual -ne $expected.ToLowerInvariant()) {
        throw "Checksum tidak cocok (unduhan rusak atau dimanipulasi)"
    }

    Expand-Archive -Path (Join-Path $tmp $archive) -DestinationPath $tmp -Force
    $exe = Join-Path $tmp "totp.exe"
    if (-not (Test-Path $exe)) {
        throw "totp.exe tidak ditemukan di dalam arsip"
    }

    $installDir = if ($env:TOTP_INSTALL_DIR) {
        $env:TOTP_INSTALL_DIR
    } else {
        Join-Path $env:LOCALAPPDATA "Programs\totp-cli"
    }
    New-Item -ItemType Directory -Path $installDir -Force | Out-Null
    Copy-Item $exe (Join-Path $installDir "totp.exe") -Force

    $userPath = [Environment]::GetEnvironmentVariable("Path", "User")
    $parts = if ($userPath) { $userPath -split ";" } else { @() }
    if ($parts -notcontains $installDir) {
        $newPath = if ($userPath) { "$installDir;$userPath" } else { $installDir }
        [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
        Write-Host "PATH pengguna ditambahkan: $installDir"
    }
    $env:Path = "$installDir;$env:Path"

    $version = & (Join-Path $installDir "totp.exe") --version
    Write-Host "Terpasang: $installDir\totp.exe"
    Write-Host "Versi: $version"
    Write-Host ""
    Write-Host "Siap digunakan. Coba: totp --version"
} finally {
    Remove-Item -Recurse -Force $tmp -ErrorAction SilentlyContinue
}
