param(
    [Parameter(Mandatory = $true)]
    [string]$ChangeId,

    [string[]]$Kinds = @("ui", "scenarios", "logic")
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))
$changeDir = Join-Path $root "changes\$ChangeId"

$templates = @{
    "ui" = @{
        Source = Join-Path $root "skills\aite2-ui-design\references\templates.md"
        Destination = "ui.md"
    }
    "scenarios" = @{
        Source = Join-Path $root "skills\aite2-scenario-design\references\templates.md"
        Destination = "scenarios.md"
    }
    "logic" = @{
        Source = Join-Path $root "skills\aite2-logic-design\references\templates.md"
        Destination = "logic.md"
    }
}

New-Item -ItemType Directory -Path $changeDir -Force | Out-Null

foreach ($kind in $Kinds) {
    if (-not $templates.ContainsKey($kind)) {
        throw "Unsupported kind: $kind. Supported kinds: $($templates.Keys -join ', ')"
    }

    $source = $templates[$kind].Source
    $destination = Join-Path $changeDir $templates[$kind].Destination

    if (-not (Test-Path $source)) {
        throw "Template not found: $source"
    }

    if (Test-Path $destination) {
        Write-Output "Skipped existing $destination"
        continue
    }

    Copy-Item -Path $source -Destination $destination
    Write-Output "Created $destination"
}

function ConvertTo-TitleCase {
    param([string]$Value)

    $parts = @()

    foreach ($segment in ($Value -replace '[-_]+', ' ' -split '\s+')) {
        if (-not $segment) {
            continue
        }

        if ($segment.Length -eq 1) {
            $parts += $segment.ToUpper()
            continue
        }

        $parts += $segment.Substring(0, 1).ToUpper() + $segment.Substring(1).ToLower()
    }

    return ($parts -join ' ')
}

$indexPath = Join-Path $changeDir "index.md"
$changeTitle = ConvertTo-TitleCase $ChangeId
$generatedDocs = Get-ChildItem -Path $changeDir -File -Filter "*.md" |
    Where-Object { $_.Name -ne "index.md" } |
    Sort-Object Name

$indexLines = @(
    "---"
    "title: $changeTitle"
    "description: Entry page for ongoing change docs"
    "---"
    ""
    "# $changeTitle"
    ""
    "This change is the entry point."
)

if ($generatedDocs.Count -gt 0) {
    $indexLines += ""
    $indexLines += "## Documents"

    foreach ($doc in $generatedDocs) {
        $name = [System.IO.Path]::GetFileNameWithoutExtension($doc.Name)
        $indexLines += "- [$name](./$name/)"
    }
} else {
    $indexLines += ""
    $indexLines += "No design documents exist yet."
}

$indexLines += ""
$indexLines += "## Next Steps"
$indexLines += '- Fill `ui.md`, `scenarios.md`, and `logic.md` in order.'
$indexLines += "- Added documents are reflected in this list automatically."

Set-Content -Path $indexPath -Value ($indexLines -join "`n") -Encoding utf8
Write-Output "Created $indexPath"
