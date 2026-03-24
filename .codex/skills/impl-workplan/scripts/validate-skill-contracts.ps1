[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$scriptRoot = Split-Path -Parent $PSCommandPath
$repoRoot = (Resolve-Path (Join-Path $scriptRoot '..\..\..\..')).Path
$skillsRoot = Join-Path $repoRoot '.codex\skills'
$agentsRoot = Join-Path $repoRoot '.codex\agents'
$changesRoot = Join-Path $repoRoot 'changes'
$packetValidator = Join-Path $repoRoot '.codex\skills\scripts\validate-packet-contracts.ps1'

$standardLoggingPrefix = '[fix-trace]'
$deprecatedTools = @('find_by_name', 'view_file', 'grep')
$legacyWorkplanFields = @('validation', 'validation_command', 'quality_gates')
$pathPatterns = @(
    '(?<![A-Za-z0-9._/-])references/[A-Za-z0-9._/-]+\.(?:md|ps1)',
    '(?<![A-Za-z0-9._/-])scripts/[A-Za-z0-9._/-]+\.(?:md|ps1)',
    '(?<![A-Za-z0-9._/-])docs/[A-Za-z0-9._/-]+\.(?:md|ps1)',
    '(?<![A-Za-z0-9._/-])\.codex/[A-Za-z0-9._/-]+\.(?:md|ps1|toml)',
    '(?<![A-Za-z0-9._/-])changes/[A-Za-z0-9._/<>\-]+\.(?:md|ps1)',
    'AGENTS\.md'
)

$findings = New-Object System.Collections.Generic.List[object]

function Get-DisplayPath {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Path
    )

    $fullPath = [System.IO.Path]::GetFullPath($Path)
    if ($fullPath.StartsWith($repoRoot, [System.StringComparison]::OrdinalIgnoreCase)) {
        return ($fullPath.Substring($repoRoot.Length).TrimStart('\', '/')) -replace '\\', '/'
    }

    return $fullPath -replace '\\', '/'
}

function Add-Finding {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Category,
        [Parameter(Mandatory = $true)]
        [string]$File,
        [Parameter(Mandatory = $true)]
        [string]$Detail,
        [int]$Line
    )

    $entry = [ordered]@{
        Category = $Category
        File = Get-DisplayPath -Path $File
        Line = $Line
        Detail = $Detail
    }

    $findings.Add([pscustomobject]$entry) | Out-Null
}

function Get-SkillRoot {
    param(
        [Parameter(Mandatory = $true)]
        [string]$SourceFile
    )

    $current = Get-Item -LiteralPath (Split-Path -Parent $SourceFile)
    while ($null -ne $current) {
        if ($current.Parent -and $current.Parent.FullName -ieq $skillsRoot) {
            return $current.FullName
        }

        $current = $current.Parent
    }

    return Split-Path -Parent $SourceFile
}

function Resolve-ReferencedPath {
    param(
        [Parameter(Mandatory = $true)]
        [string]$SourceFile,
        [Parameter(Mandatory = $true)]
        [string]$Reference
    )

    if ($Reference -match '<|>' -or $Reference.Contains('...')) {
        return $null
    }

    $normalized = $Reference -replace '/', '\'
    if ($Reference -eq 'AGENTS.md') {
        return Join-Path $repoRoot $normalized
    }

    if ($Reference -match '^(references|scripts)/') {
        return Join-Path (Get-SkillRoot -SourceFile $SourceFile) $normalized
    }

    if ($Reference -match '^(docs|changes|\.codex)/') {
        return Join-Path $repoRoot $normalized
    }

    return $null
}

function Test-ReferencedFiles {
    param(
        [Parameter(Mandatory = $true)]
        [System.IO.FileInfo[]]$Files
    )

    foreach ($file in $Files) {
        $seenReferences = @{}
        $lines = Get-Content -LiteralPath $file.FullName
        for ($lineIndex = 0; $lineIndex -lt $lines.Count; $lineIndex++) {
            $line = $lines[$lineIndex]
            foreach ($pattern in $pathPatterns) {
                foreach ($match in [regex]::Matches($line, $pattern, [System.Text.RegularExpressions.RegexOptions]::IgnoreCase)) {
                    $reference = $match.Value
                    if ($seenReferences.ContainsKey($reference)) {
                        continue
                    }

                    $seenReferences[$reference] = $true
                    $resolvedPath = Resolve-ReferencedPath -SourceFile $file.FullName -Reference $reference
                    if ($null -eq $resolvedPath) {
                        continue
                    }

                    if (-not (Test-Path -LiteralPath $resolvedPath)) {
                        Add-Finding -Category 'missing_reference' -File $file.FullName -Line ($lineIndex + 1) -Detail ("missing referenced file: {0}" -f $reference)
                    }
                }
            }
        }
    }
}

function Test-DeprecatedTools {
    param(
        [Parameter(Mandatory = $true)]
        [System.IO.FileInfo[]]$Files
    )

    $pattern = '\b(?:' + (($deprecatedTools | ForEach-Object { [regex]::Escape($_) }) -join '|') + ')\b'
    foreach ($file in $Files) {
        $lines = Get-Content -LiteralPath $file.FullName
        for ($lineIndex = 0; $lineIndex -lt $lines.Count; $lineIndex++) {
            $line = $lines[$lineIndex]
            foreach ($match in [regex]::Matches($line, $pattern, [System.Text.RegularExpressions.RegexOptions]::IgnoreCase)) {
                Add-Finding -Category 'deprecated_tool_name' -File $file.FullName -Line ($lineIndex + 1) -Detail ("deprecated tool name: {0}" -f $match.Value)
            }
        }
    }
}

function Test-LoggingPrefix {
    param(
        [Parameter(Mandatory = $true)]
        [System.IO.FileInfo[]]$Files
    )

    $pattern = '\[fix-[^\]]+\]'
    foreach ($file in $Files) {
        $lines = Get-Content -LiteralPath $file.FullName
        for ($lineIndex = 0; $lineIndex -lt $lines.Count; $lineIndex++) {
            $line = $lines[$lineIndex]
            foreach ($match in [regex]::Matches($line, $pattern, [System.Text.RegularExpressions.RegexOptions]::IgnoreCase)) {
                if ($match.Value -ieq $standardLoggingPrefix) {
                    continue
                }

                Add-Finding -Category 'logging_prefix_drift' -File $file.FullName -Line ($lineIndex + 1) -Detail ("expected {0}, found {1}" -f $standardLoggingPrefix, $match.Value)
            }
        }
    }
}

function Test-LegacyWorkplanFields {
    param(
        [Parameter(Mandatory = $true)]
        [System.IO.FileInfo[]]$Files
    )

    $pattern = '^\s*-\s*(?:' + (($legacyWorkplanFields | ForEach-Object { [regex]::Escape($_) }) -join '|') + ')\s*:'
    foreach ($file in $Files) {
        $lines = Get-Content -LiteralPath $file.FullName
        for ($lineIndex = 0; $lineIndex -lt $lines.Count; $lineIndex++) {
            $line = $lines[$lineIndex]
            if ($line -match $pattern) {
                Add-Finding -Category 'workplan_field_mismatch' -File $file.FullName -Line ($lineIndex + 1) -Detail ("legacy field: {0}" -f $Matches[0].Trim())
            }
        }
    }
}

function Test-ValidationCommandsPresence {
    param(
        [Parameter(Mandatory = $true)]
        [System.IO.FileInfo[]]$Files
    )

    foreach ($file in $Files) {
        $lines = Get-Content -LiteralPath $file.FullName
        for ($lineIndex = 0; $lineIndex -lt $lines.Count; $lineIndex++) {
            if ($lines[$lineIndex] -notmatch '^(?<indent>\s*)-\s*section_id\s*:') {
                continue
            }

            $sectionIndent = $Matches['indent'].Length
            $hasValidationCommands = $false
            $looksLikePlanSection = $false

            for ($scanIndex = $lineIndex + 1; $scanIndex -lt $lines.Count; $scanIndex++) {
                $scanLine = $lines[$scanIndex]

                if ($scanLine -match '^```' -or $scanLine -match '^\s*##\s+') {
                    break
                }

                if ($scanLine -match '^(?<indent>\s*)-\s*section_id\s*:' -and $Matches['indent'].Length -le $sectionIndent) {
                    break
                }

                if ($scanLine -match '^(?<indent>\s*)-\s*\[[ xX]\]\s+' -and $Matches['indent'].Length -le $sectionIndent) {
                    break
                }

                if ($scanLine -match '^\s*(?:-\s*)?title\s*:' -or $scanLine -match '^\s*(?:-\s*)?owner\s*:') {
                    $looksLikePlanSection = $true
                }

                if ($scanLine -match '^\s*(?:-\s*)?validation_commands\s*:') {
                    $hasValidationCommands = $true
                    break
                }
            }

            if ($looksLikePlanSection -and -not $hasValidationCommands) {
                Add-Finding -Category 'workplan_field_mismatch' -File $file.FullName -Line ($lineIndex + 1) -Detail 'section schema is missing validation_commands'
            }
        }
    }
}

$skillMarkdownFiles = @(Get-ChildItem -LiteralPath $skillsRoot -Recurse -Filter '*.md' -File)
$skillFiles = @($skillMarkdownFiles | Where-Object { $_.Name -eq 'SKILL.md' })
$agentFiles = @(Get-ChildItem -LiteralPath $agentsRoot -Filter '*.toml' -File)
$taskFiles = @()
if (Test-Path -LiteralPath $changesRoot) {
    $taskFiles = @(Get-ChildItem -LiteralPath $changesRoot -Recurse -Filter 'tasks.md' -File)
}

Test-ReferencedFiles -Files $skillMarkdownFiles
Test-DeprecatedTools -Files ($skillMarkdownFiles + $agentFiles)
Test-LoggingPrefix -Files ($skillFiles + $agentFiles)
Test-LegacyWorkplanFields -Files ($skillMarkdownFiles + $taskFiles)
Test-ValidationCommandsPresence -Files ($skillMarkdownFiles + $taskFiles)

$sortedFindings = $findings | Sort-Object Category, File, Line, Detail
$packetValidatorExitCode = 0

if (Test-Path -LiteralPath $packetValidator) {
    Write-Host 'Packet contract validation (delegated)'
    & powershell -ExecutionPolicy Bypass -File $packetValidator
    $packetValidatorExitCode = $LASTEXITCODE
}

Write-Host 'Skill contract validation'
Write-Host ("- checked_skill_markdown: {0}" -f $skillMarkdownFiles.Count)
Write-Host ("- checked_agents: {0}" -f $agentFiles.Count)
Write-Host ("- checked_tasks: {0}" -f $taskFiles.Count)

if ($sortedFindings.Count -eq 0 -and $packetValidatorExitCode -eq 0) {
    Write-Host '- findings: 0'
    exit 0
}

Write-Host ("- findings: {0}" -f $sortedFindings.Count)
foreach ($finding in $sortedFindings) {
    $location = if ($finding.Line) {
        '{0}:{1}' -f $finding.File, $finding.Line
    }
    else {
        $finding.File
    }

    Write-Host ("[{0}] {1} - {2}" -f $finding.Category, $location, $finding.Detail)
}

exit 1
