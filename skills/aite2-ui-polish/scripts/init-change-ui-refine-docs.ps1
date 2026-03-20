param(
    [Parameter(Mandatory = $true)]
    [string]$ChangeId
)

$ErrorActionPreference = "Stop"

$root = Split-Path -Parent (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))
$changeDir = Join-Path $root "changes\$ChangeId"
$contextBoardDir = Join-Path $changeDir "context_board"

$boardTemplates = @(
    @{
        Source = Join-Path $root "skills\aite2-ui-polish\references\context-board\current_context.md"
        Destination = Join-Path $contextBoardDir "current_context.md"
    }
    @{
        Source = Join-Path $root "skills\aite2-ui-polish\references\context-board\handoff.md"
        Destination = Join-Path $contextBoardDir "handoff.md"
    }
    @{
        Source = Join-Path $root "skills\aite2-ui-polish\references\context-board\findings.md"
        Destination = Join-Path $contextBoardDir "findings.md"
    }
)

New-Item -ItemType Directory -Path $changeDir -Force | Out-Null
New-Item -ItemType Directory -Path $contextBoardDir -Force | Out-Null

foreach ($boardTemplate in $boardTemplates) {
    if (-not (Test-Path $boardTemplate.Source)) {
        throw "Template not found: $($boardTemplate.Source)"
    }

    if (Test-Path $boardTemplate.Destination) {
        Write-Output "Skipped existing $($boardTemplate.Destination)"
        continue
    }

    Copy-Item -Path $boardTemplate.Source -Destination $boardTemplate.Destination
    Write-Output "Created $($boardTemplate.Destination)"
}

$indexPath = Join-Path $changeDir "index.md"

$indexLines = @(
    "---"
    "title: $ChangeId"
    "description: Entry page for UI refine work"
    "---"
    ""
    "# $ChangeId"
    ""
    "This change is the entry point for UI refine work."
    ""
    "## Next Steps"
    "- Use `context_board/` for observation, fix plan, and review handoff."
    "- Keep logic changes out of this change unless explicitly expanded."
)

Set-Content -Path $indexPath -Value ($indexLines -join "`n") -Encoding utf8
Write-Output "Created $indexPath"
