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
        Source = Join-Path $root "skills\fix-direction\references\context-board\current_context.md"
        Destination = Join-Path $contextBoardDir "current_context.md"
    }
    @{
        Source = Join-Path $root "skills\fix-direction\references\context-board\handoff.md"
        Destination = Join-Path $contextBoardDir "handoff.md"
    }
    @{
        Source = Join-Path $root "skills\fix-direction\references\context-board\findings.md"
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
    "description: Entry page for bugfix investigation"
    "---"
    ""
    "# $ChangeId"
    ""
    "This change is the entry point for bugfix work."
    ""
    "## Next Steps"
    "- Use `context_board/` for repro, cause investigation handoff, and findings."
    "- Record cause narrowing before asking `fix-work` to change code."
)

Set-Content -Path $indexPath -Value ($indexLines -join "`n") -Encoding utf8
Write-Output "Created $indexPath"
