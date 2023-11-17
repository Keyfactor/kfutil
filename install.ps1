# Copyright 2023 The Keyfactor Command Authors.
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

<#
.SYNOPSIS
This script installs the Keyfactor Command Utility (kfutil).

.DESCRIPTION
This script downloads and installs the Keyfactor Command Utility. It can be run
with or without administrator privileges. If run without administrator privileges,
the script will install the utility into the user's home directory. If run with
administrator privileges, the script can install the utility into a custom directory.

.PARAMETER Uninstall
Indicates that the script should perform an uninstallation.

.PARAMETER BinaryName
The name of the binary to install. Defaults to "kfutil". Don't change this unless
you know what you're doing.

.PARAMETER Version
The version of the binary to install. If not specified, the script will install
the latest stable release. If specified, the script will install the specified
version. If the specified version is not a stable release, the script will warn
the user that the version is not stable.

.PARAMETER InstallDir
The directory to install the binary into. If not specified, the script will
install the binary into the user's home directory. If specified, the script will
install the binary into the specified directory, but will require administrator
privileges to do so if the directory is not in the user's home directory.

.PARAMETER VerifyChecksum
Indicates that the script should verify the SHA256 checksum of the downloaded
binary. Defaults to $true.

.EXAMPLE
PS> .\install.ps1
Installs the latest stable release of the Keyfactor Command Utility into the
user's home directory.

.EXAMPLE
PS> .\install.ps1 -Version v1.0.0 -InstallDir C:\Windows\System32\WindowsPowerShell\v1.0
Installs version 1.0.0 of the Keyfactor Command Utility into the specified directory.

.EXAMPLE
PS> .\install.ps1 -Uninstall
Uninstalls the Keyfactor Command Utility from the directory in the PATH.

.NOTES
Additional information about the script.

.LINK
https://github.com/Keyfactor/kfutil

#>

param(
    [string]$BinaryName = "kfutil",
    [string]$Version,
    [string]$InstallDir = [System.IO.Path]::Combine([System.Environment]::GetFolderPath("UserProfile"), "AppData", "Local", "Microsoft", "WindowsApps"),
    [bool]$VerifyChecksum = $true,
    [switch]$Uninstall
)

function Get-Architecture {
    param( [bool]$IsPowershellCore )
    # If script is running on PowerShell Core host, use uname
    if ($IsPowershellCore) {
        $ARCH = uname -m

        switch -Wildcard ($ARCH) {
            "armv5*" { $ARCH = "armv5" }
            "armv6*" { $ARCH = "armv6" }
            "armv7*" { $ARCH = "arm" }
            "aarch64" { $ARCH = "arm64" }
            "arm64" { $ARCH = "arm64" }
            "x86" { $ARCH = "386" }
            "x86_64" { $ARCH = "amd64" }
            "i686" { $ARCH = "386" }
            "i386" { $ARCH = "386" }
        }

        return $ARCH
    }

    $ARCH = (Get-WmiObject Win32_Processor).Architecture

    switch ($ARCH) {
        0 { $ARCH = "386" }       # x86
        5 { $ARCH = "arm" }       # ARM
        9 { $ARCH = "amd64" }     # x64
        12 { $ARCH = "arm64" }    # ARM64
        default { $ARCH = "unknown" }
    }

    # Return the architecture
    return $ARCH
}

function Get-OperatingSystem {
    param ( [bool]$IsPowershellCore )
    [string]$OS = $null

    if ($IsPowershellCore) {
        # If script is running on PowerShell Core host, use Automatic Variables

        if ($IsLinux) {
            $OS = "linux"
        }
        elseif ($IsMacOS) {
            $OS = "darwin"
        }
        elseif ($IsWindows) {
            $OS = "windows"
        }
    } else {
        # Otherwise, ask .NET for the OS description

        if ([System.Environment]::OSVersion.Platform -eq "Win32NT") {
            $OS = "windows"
        }
    }
    
    # If we couldn't determine the OS, fail the installation
    if ([string]::IsNullOrWhiteSpace($OS)) {
        throw "Unable to determine operating system"
    }

    # Return the OS
    return $OS
}

function Test-SupportedHost {
    param (
        [string]$OS,
        [string]$Architecture
    )

    $SupportedBuilds = @(
        "darwin-amd64",
        "darwin-arm64",
        "linux-386",
        "linux-amd64",
        "linux-arm",
        "linux-arm64",
        "linux-ppc64le",
        "linux-s390x",
        "windows-386",
        "windows-amd64",
        "windows-arm",
        "windows-arm64"
    )

    $matchFound = $false
    foreach ($build in $supportedBuilds) {
        if ("${build}" -eq "${OS}-${Architecture}") {
            $matchFound = $true
            break
        }
    }

    if (!$matchFound) {
        throw "Unsupported operating system and architecture combination: ${OS}-${Architecture}"
    }
}

function Get-InstallVersion {
    param (
        [string]$Version
    )

    [string]$SemVer = $null

    $RemoteReleaseUrl = "https://api.github.com/repos/keyfactor/$BinaryName/releases"
    $Releases = (Invoke-WebRequest -UseBasicParsing $RemoteReleaseUrl | ConvertFrom-Json)

    if (![string]::IsNullOrWhiteSpace($Version)) {
        # Strip the leading 'v' from the version string
        $SemVer = $Version -replace '^v',''

        # Verify that the version exists as a release before continuing
        $Release = $Releases | Where-Object { $_.tag_name -eq "v$SemVer" }
        if ([string]::IsNullOrWhiteSpace($Release.tag_name)) {
            throw "Cannot find release $Version for $BinaryName"
        }

        Write-Host "$BinaryName version $Version exists"
    } else {
        # If no version was specified, get the latest release
        $LatestRelease = $Releases | Where-Object { -not $_.prerelease } | Sort-Object -Property created_at -Descending | Select-Object -First 1
        $SemVer = $LatestRelease.tag_name -replace '^v',''
    }

    return $SemVer
}

function Test-IsVersionInstalled {
    param (
        [string]$BinaryName,
        [string]$Version
    )

    try {
        $BinaryPath = (Get-Command $BinaryName -ErrorAction Stop).Path

        $InstalledVersion = [string](& $BinaryPath version).Split(' ')[-1] -replace '^v',''
        if ($InstalledVersion -eq $Version) {
            Write-Host "$BinaryName version $Version is already installed"
            return $true
        } else {
            Write-Host "Changing from $BinaryName version $InstalledVersion to $Version"
            return $false
        }
    }
    catch {
        return $false
    }

    return $false
}

function New-TempInstallDir {
    # Get the base temporary folder
    $BaseTempPath = [System.IO.Path]::GetTempPath()

    # Create a unique directory name with a prefix
    $UniqueTempDir = [System.IO.Path]::Combine($BaseTempPath, "$BinaryName-installer-" + [Guid]::NewGuid().ToString())

    # Create the directory
    New-Item -ItemType Directory -Path $UniqueTempDir | Out-Null

    # Return the path
    return $UniqueTempDir
}

function Get-BinaryAndSum {
    param (
        [string]$BaseTempDir,
        [string]$BinaryName,
        [string]$Version,
        [string]$OS,
        [string]$Architecture
    )

    $ReleaseUrl = "https://github.com/Keyfactor/${BinaryName}/releases/download/v${Version}"
    $Dist = "${BinaryName}_${Version}_${OS}_${Architecture}.zip"
    
    $DistUrl = "${ReleaseUrl}/${Dist}"
    $SumUrl = "${ReleaseUrl}/${BinaryName}_${Version}_SHA256SUMS"

    Write-Host "Downloading $BinaryName $Version for $OS-$Architecture"

    # Download the binary and checksum
    $DistPath = [System.IO.Path]::Combine($BaseTempDir, $Dist)
    $SumPath = [System.IO.Path]::Combine($BaseTempDir, "${BinaryName}.sum")

    Invoke-WebRequest -UseBasicParsing -OutFile $DistPath $DistUrl -ErrorAction Stop
    Invoke-WebRequest -UseBasicParsing -OutFile $SumPath $SumUrl  -ErrorAction Stop

    return @{
        dist = $Dist
        dist_path = $DistPath
        sum_path = $SumPath
    }
}

function Test-Checksum {
    param (
        [string]$Dist,
        [string]$DistPath,
        [string]$SumPath
    )

    if (-not $VerifyChecksum) {
        Write-Host "Skipping checksum verification"
        return
    }

    Write-Host "Verifying checksum... " -NoNewline

    # Extract the expected checksum from the SHA256SUMS file
    $ExpectedSum = [string](Get-Content $SumPath | Select-String -Pattern $Dist | Select-Object -First 1 | ForEach-Object { $_.ToString().Split(' ')[0] }).ToUpper()
    
    # Calculate the actual checksum of the binary
    $ActualSum = [string](Get-FileHash $DistPath -Algorithm SHA256 | Select-Object -ExpandProperty Hash).ToUpper()

    if ($ExpectedSum -ne $ActualSum) {
        throw "SHA sum of $BinaryName binary does not match. Expected $ExpectedSum, got $ActualSum. Aborting."
    }

    Write-Host "Done."
}

function Install-Binary {
    param (
        [string]$BaseTempDir,
        [string]$DistPath,
        [string]$InstallDir,
        [string]$BinaryName,
        [string]$Os
    )

    Write-Host "Preparing to install $BinaryName into ${InstallDir}"

    $TempBinDir = [System.IO.Path]::Combine($BaseTempDir, "bin")

    # Unzip the binary to the temp directory
    Expand-Archive -Path $DistPath -DestinationPath $TempBinDir

    # Create the install directory if it doesn't exist
    if (-not (Test-Path $InstallDir)) {
        New-Item -ItemType Directory -Path $InstallDir | Out-Null
    }

    # Adjust the binary name if we're on Windows
    if ($Os -eq "windows") {
        $BinaryName = $BinaryName + ".exe"
    }

    # Copy the binary to the install directory
    $TempBinDir = [System.IO.Path]::Combine($TempBinDir, $BinaryName)
    Copy-Item -Path $TempBinDir -Destination $InstallDir -Recurse -Force -ErrorAction Stop

    Write-Host "$BinaryName installed into $([System.IO.Path]::Combine(${InstallDir},$BinaryName))"
}

function Uninstall-Binary {
    param (
        [string]$BinaryName
    )

    try {
        Get-Command $BinaryName -ErrorAction Stop | Out-Null
    }
    catch {
        Write-Warning "$BinaryName is not installed."
        return
    }
    
    $BinaryPath = (Get-Command $BinaryName).Path

    Write-Host "Uninstalling $BinaryName from $BinaryPath... " -NoNewline

    # Uninstall binary
    Remove-Item -Path $BinaryPath -Recurse -Force

    
    try {
        Get-Command $BinaryName -ErrorAction Stop
    }
    catch {
        Write-Host "Done."
        return
    }

    throw "$BinaryName is still installed. Uninstallation failed."
}

function Test-InstalledVersion {
    param (
        [string]$BinaryName,
        [string]$Version,
        [string]$InstallDir
    )

    try {
        Get-Command $BinaryName -ErrorAction Stop | Out-Null
    }
    catch {
        throw "$BinaryName not found. Is $InstallDir in your PATH?"
    }

    $InstalledVersion = [string](& $BinaryName version).Split(' ')[-1] -replace '^v',''
    if ($InstalledVersion -eq $Version) {
        Write-Host "$BinaryName $Version is installed and available."
    } else {
        # If the Version is an RC or otherwise prerelease, we'll allow it to be installed
        # but we'll warn the user that it's not the latest stable version
        if ($Version -match "-" -and $Version.Split('-')[0] -eq $InstalledVersion) {
            Write-Warning "$BinaryName $Version is installed, but is not an official release. Behavior may be unstable; use at your own risk."
            return
        }

        throw "Installed version of $BinaryName ($InstalledVersion) does not match requested version ($Version)"
    }

    return
}

function Clear-InstallArtifacts {
    param (
        [string]$BaseTempDir
    )

    if ([string]::IsNullOrWhiteSpace($BaseTempDir)) {
        return
    }

    Write-Host "Cleaning up temporary files... " -NoNewline

    # Remove the temporary directory
    Remove-Item -Path $BaseTempDir -Recurse -Force

    Write-Host "Done."
}

try
{
    if ($Uninstall) {
        Uninstall-Binary -BinaryName $BinaryName
        exit 0
    }

    $isPowershellCore = $PSVersionTable.PSEdition -eq "Core"
    $UniqueTempDir = $null
    
    # Get the architecture and operating system
    $Architecture = Get-Architecture -IsPowershellCore $isPowershellCore
    $Os = Get-OperatingSystem -IsPowershellCore $isPowershellCore

    # Verify that the host is supported, and deps on the host are met
    Test-SupportedHost -OS $Os -Architecture $Architecture

    # Verify or get the version to install
    $Version = Get-InstallVersion -Version $Version

    # Check if the binary is already installed
    if (-Not (Test-IsVersionInstalled -BinaryName $BinaryName -Version $Version)) {
        $UniqueTempDir = New-TempInstallDir
        $files = Get-BinaryAndSum -BaseTempDir $UniqueTempDir -BinaryName $BinaryName -Version $Version -OS $Os -Architecture $Architecture
        Test-Checksum -Dist $files.dist -DistPath $files.dist_path -SumPath $files.sum_path
        Install-Binary -BaseTempDir $UniqueTempDir -DistPath $files.dist_path -InstallDir $InstallDir -BinaryName $BinaryName -Os $Os

        Test-InstalledVersion -BinaryName $BinaryName -Version $Version -InstallDir $InstallDir
    }

    Clear-InstallArtifacts -BaseTempDir $UniqueTempDir
}
catch
{
    Write-Output $_
    Clear-InstallArtifacts -BaseTempDir $UniqueTempDir
    exit 1
}
