$ErrorActionPreference = "Stop"

$Repo = if ($env:MCP_PB_REPO) { $env:MCP_PB_REPO } else { "mreyeswilson/pocketmcp" }
$BinaryName = "pocketmcp"

if ($env:MCP_PB_VERSION) {
  $Tag = $env:MCP_PB_VERSION
} else {
  $Latest = Invoke-RestMethod -Uri "https://api.github.com/repos/$Repo/releases/latest"
  $Tag = $Latest.tag_name
}

if (-not $Tag) {
  throw "Failed to resolve release tag from $Repo."
}

$Arch = $env:PROCESSOR_ARCHITECTURE
if ($Arch -notin @("AMD64", "x86_64")) {
  throw "Unsupported architecture: $Arch. Supported: x86_64."
}

$Target = "x86_64-pc-windows-msvc"
$Asset = "$BinaryName-$Tag-$Target.exe"
$Url = "https://github.com/$Repo/releases/download/$Tag/$Asset"

$InstallDir = Join-Path $env:LOCALAPPDATA "mcp-pocketbase-admin\bin"
New-Item -ItemType Directory -Force -Path $InstallDir | Out-Null

$Destination = Join-Path $InstallDir "$BinaryName.exe"
$TempFile = Join-Path $env:TEMP ("$BinaryName-" + [Guid]::NewGuid().ToString() + ".exe")

Write-Host "Downloading $Url"
Invoke-WebRequest -Uri $Url -OutFile $TempFile
Move-Item -Force -Path $TempFile -Destination $Destination

Write-Host "Installed $BinaryName $Tag to $Destination"
Write-Host "Next steps:"
Write-Host "  1) Add $InstallDir to your PATH if needed"
Write-Host "  2) Run: $BinaryName serve --url <url> --email <email> --password <password>"
Write-Host "  3) Or run: $BinaryName install --client all --url <url> --email <email> --password <password>"
