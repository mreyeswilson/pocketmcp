param(
  [string] $Version,
  [string] $RepoOverride
)

$ErrorActionPreference = "Stop"

$Repo = if ($RepoOverride) {
  $RepoOverride
} elseif ($env:MCP_PB_REPO) {
  $env:MCP_PB_REPO
} else {
  "mreyeswilson/pocketmcp"
}

$BinaryBaseName = "pocketmcp"
$BinaryFileName = "pocketmcp.exe"

try {
  $tls12 = [Net.SecurityProtocolType]::Tls12
  $tls13 = if ([Enum]::GetNames([Net.SecurityProtocolType]) -contains "Tls13") {
    [Net.SecurityProtocolType]::Tls13
  } else {
    0
  }
  [Net.ServicePointManager]::SecurityProtocol = $tls12 -bor $tls13
} catch {
  [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
}

function Download-WithRetry {
  param(
    [Parameter(Mandatory = $true)] [string] $Url,
    [Parameter(Mandatory = $true)] [string] $OutFile,
    [int] $MaxAttempts = 3
  )

  $hasBits = $null -ne (Get-Command Start-BitsTransfer -ErrorAction SilentlyContinue)
  $lastErrors = @()

  for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
    if (Test-Path $OutFile) {
      Remove-Item -Force -Path $OutFile
    }

    $attemptErrors = @()

    try {
      Invoke-WebRequest -Uri $Url -OutFile $OutFile
      return
    } catch {
      $attemptErrors += "Invoke-WebRequest failed: $($_.Exception.Message)"
    }

    if ($hasBits) {
      try {
        Start-BitsTransfer -Source $Url -Destination $OutFile -TransferType Download
        return
      } catch {
        $attemptErrors += "Start-BitsTransfer failed: $($_.Exception.Message)"
      }
    }

    $lastErrors = $attemptErrors
    if ($attempt -lt $MaxAttempts) {
      $delaySeconds = [int][Math]::Pow(2, $attempt)
      Write-Warning "Download attempt $attempt/$MaxAttempts failed. Retrying in $delaySeconds seconds..."
      Start-Sleep -Seconds $delaySeconds
    }
  }

  throw "Failed to download asset after $MaxAttempts attempts.`n$($lastErrors -join "`n")"
}

function Get-LatestTag {
  param(
    [Parameter(Mandatory = $true)] [string] $Repository
  )

  try {
    $latestRelease = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repository/releases/latest"
    if ($latestRelease.tag_name) {
      return $latestRelease.tag_name
    }
  } catch {
  }

  $latestTag = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repository/tags?per_page=1"
  if ($latestTag -and $latestTag[0].name) {
    return $latestTag[0].name
  }

  return $null
}

if ($Version) {
  $Tag = $Version
} elseif ($env:VERSION) {
  $Tag = $env:VERSION
} elseif ($env:MCP_PB_VERSION) {
  $Tag = $env:MCP_PB_VERSION
} else {
  $Tag = Get-LatestTag -Repository $Repo
}

if (-not $Tag) {
  throw "Failed to resolve release tag from $Repo."
}

$Arch = $env:PROCESSOR_ARCHITECTURE
if ($Arch -notin @("AMD64", "x86_64")) {
  throw "Unsupported architecture: $Arch. Supported: x86_64."
}

$Target = "x86_64-pc-windows-msvc"
$Asset = "$BinaryBaseName-$Tag-$Target.exe"
$Url = "https://github.com/$Repo/releases/download/$Tag/$Asset"

$InstallDir = Join-Path $env:LOCALAPPDATA "pocketmcp\bin"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$Destination = Join-Path $InstallDir $BinaryFileName
$TempFile = Join-Path $env:TEMP ("$BinaryBaseName-" + [Guid]::NewGuid().ToString() + ".exe")

Write-Host "Downloading $Url"
Download-WithRetry -Url $Url -OutFile $TempFile -MaxAttempts 3
Move-Item -Force -Path $TempFile -Destination $Destination

Write-Host "Installed $BinaryFileName $Tag to $Destination"
Write-Host "Next steps:"
Write-Host "  1) Add $InstallDir to your PATH if needed"
Write-Host "  2) Run: $BinaryFileName serve --url <url> --email <email> --password <password>"
Write-Host "  3) Or run: $BinaryFileName install --client all --url <url> --email <email> --password <password>"
Write-Host "Version selection: latest by default; set VERSION (or MCP_PB_VERSION) or pass -Version <tag>."
