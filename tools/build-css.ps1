# build-css.ps1 — équivalent Windows de tools/build-css.sh : régénère
# web/static/tailwind.css (classes Tailwind utilisées dans les sources Go +
# palette de tailwind.config.js) puis embarqué via go:embed (web/assets.go).
#
# Usage  : powershell -File tools/build-css.ps1
# Watch  : $env:CSS_WATCH = "1" avant le lancement (mode dev).
# Dépendance : CLI Tailwind standalone (pas de Node), version épinglée,
# mise en cache dans $HOME\.cache\grz (même cache que le .sh).

Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

$TTY_VERSION = "v3.4.17"
$CACHE = Join-Path $HOME ".cache\grz"
$REPO_ROOT = (Resolve-Path (Join-Path $PSScriptRoot "..")).Path

New-Item -ItemType Directory -Force -Path $CACHE | Out-Null
$BIN = Join-Path $CACHE "tailwindcss.exe"

if (-not (Test-Path $BIN)) {
    Write-Host "Téléchargement de tailwindcss $TTY_VERSION (CLI standalone Windows)…"
    switch ($env:PROCESSOR_ARCHITECTURE) {
        "AMD64" { $ASSET = "tailwindcss-windows-x64.exe" }
        "ARM64" { $ASSET = "tailwindcss-windows-arm64.exe" }
        default { throw "Architecture non supportée : $env:PROCESSOR_ARCHITECTURE" }
    }
    $url = "https://github.com/tailwindlabs/tailwindcss/releases/download/$TTY_VERSION/$ASSET"
    Invoke-WebRequest -UseBasicParsing -Uri $url -OutFile $BIN
}

Push-Location $REPO_ROOT
try {
    $tailArgs = @("-c", "tailwind.config.js", "-i", "web/static/input.css",
                  "-o", "web/static/tailwind.css", "--minify")
    if ("$env:CSS_WATCH" -eq "1") {
        $tailArgs = $tailArgs[0..($tailArgs.Count - 2)] + @("--watch")
    }
    & $BIN @tailArgs
} finally {
    Pop-Location
}
