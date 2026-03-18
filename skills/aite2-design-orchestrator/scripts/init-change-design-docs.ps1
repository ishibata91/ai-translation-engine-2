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
        Source = Join-Path $root "skills\aite2-ui-design\references\ui-contract-template.md"
        Destination = "ui.md"
    }
    "scenarios" = @{
        Source = Join-Path $root "skills\aite2-scenario-design\references\scenario-template.md"
        Destination = "scenarios.md"
    }
    "logic" = @{
        Source = Join-Path $root "skills\aite2-logic-design\references\logic-template.md"
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
        throw "Destination already exists: $destination"
    }

    Copy-Item -Path $source -Destination $destination
    Write-Output "Created $destination"
}
