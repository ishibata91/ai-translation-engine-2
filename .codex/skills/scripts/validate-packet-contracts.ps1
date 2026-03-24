[CmdletBinding()]
param(
    [string[]]$Roots
)

$ErrorActionPreference = 'Stop'

$scriptRoot = Split-Path -Parent $PSCommandPath
$repoRoot = (Resolve-Path (Join-Path $scriptRoot '..\..\..')).Path

if (-not $Roots -or $Roots.Count -eq 0) {
    $defaultChangesRoot = Join-Path $repoRoot 'changes'
    if (Test-Path -LiteralPath $defaultChangesRoot) {
        $Roots = @($defaultChangesRoot)
    }
    else {
        $Roots = @()
    }
}

$packetSchemas = [ordered]@{
    'impl-distill.packet.json' = @('invoked_skill', 'invoked_by', 'change', 'task', 'known_facts', 'unknowns', 'current_scope', 'next_action')
    'fix-distill.packet.json' = @('invoked_skill', 'invoked_by', 'change', 'symptom', 'known_facts', 'unknowns', 'current_scope', 'next_action')
    'impl-workplan.packet.json' = @('change', 'tasks_path', 'progress_snapshot', 'shared_contracts', 'dispatch_order', 'sections', 'unresolved')
    'fix-trace.packet.json' = @('invoked_skill', 'invoked_by', 'change', 'current_hypothesis', 'unknowns', 'current_scope', 'next_action')
    'impl-review.feedback.json' = @('score', 'severity', 'location', 'affected_sections', 'violated_contract', 'required_delta', 'recheck', 'docs_sync_needed')
    'fix-review.feedback.json' = @('score', 'severity', 'location', 'violated_contract', 'required_delta', 'recheck', 'docs_sync_needed')
}

$validationArtifactFields = @('target_artifact', 'validator', 'valid', 'errors', 'retry_count')
$unexpectedFieldsByPacket = @{
    'impl-review.feedback.json' = @('reviewer_actions')
}

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
        [string]$Detail
    )

    $findings.Add([pscustomobject]@{
            Category = $Category
            File = Get-DisplayPath -Path $File
            Detail = $Detail
        }) | Out-Null
}

function ConvertFrom-JsonObject {
    param(
        [Parameter(Mandatory = $true)]
        [string]$File
    )

    try {
        $raw = Get-Content -LiteralPath $File -Raw
        if ([string]::IsNullOrWhiteSpace($raw)) {
            Add-Finding -Category 'invalid_json' -File $File -Detail 'empty file'
            return $null
        }

        $parsed = $raw | ConvertFrom-Json
        if ($null -eq $parsed) {
            Add-Finding -Category 'invalid_json' -File $File -Detail 'parsed to null'
            return $null
        }

        return $parsed
    }
    catch {
        Add-Finding -Category 'invalid_json' -File $File -Detail $_.Exception.Message
        return $null
    }
}

function Test-JsonRootContextBoard {
    param(
        [Parameter(Mandatory = $true)]
        [string]$File
    )

    return $File -match '(?:^|[\\/])changes[\\/].+[\\/]context_board[\\/]'
}

function Test-RequiredFields {
    param(
        [Parameter(Mandatory = $true)]
        [string]$File,
        [Parameter(Mandatory = $true)]
        [object]$Object,
        [Parameter(Mandatory = $true)]
        [string[]]$RequiredFields
    )

    foreach ($field in $RequiredFields) {
        if (-not $Object.PSObject.Properties.Name.Contains($field)) {
            Add-Finding -Category 'missing_required_field' -File $File -Detail ("missing field: {0}" -f $field)
            continue
        }

        $value = $Object.$field
        if ($null -eq $value) {
            Add-Finding -Category 'missing_required_field' -File $File -Detail ("null field: {0}" -f $field)
        }
    }
}

function Test-ValidationArtifact {
    param(
        [Parameter(Mandatory = $true)]
        [string]$File,
        [Parameter(Mandatory = $true)]
        [object]$Object
    )

    Test-RequiredFields -File $File -Object $Object -RequiredFields $validationArtifactFields

    if ($Object.PSObject.Properties.Name.Contains('valid') -and $Object.valid -isnot [bool]) {
        Add-Finding -Category 'invalid_validation_artifact' -File $File -Detail 'valid must be boolean'
    }

    if ($Object.PSObject.Properties.Name.Contains('errors') -and $Object.errors -isnot [System.Collections.IEnumerable]) {
        Add-Finding -Category 'invalid_validation_artifact' -File $File -Detail 'errors must be an array'
    }

    if ($Object.PSObject.Properties.Name.Contains('retry_count')) {
        $retryCount = $Object.retry_count
        if ($retryCount -isnot [int] -and $retryCount -isnot [long]) {
            Add-Finding -Category 'invalid_validation_artifact' -File $File -Detail 'retry_count must be integer'
        }
        elseif ($retryCount -lt 0) {
            Add-Finding -Category 'invalid_validation_artifact' -File $File -Detail 'retry_count must be >= 0'
        }
    }
}

function Test-PacketArtifact {
    param(
        [Parameter(Mandatory = $true)]
        [string]$File,
        [Parameter(Mandatory = $true)]
        [object]$Object,
        [Parameter(Mandatory = $true)]
        [string]$PacketName
    )

    Test-RequiredFields -File $File -Object $Object -RequiredFields $packetSchemas[$PacketName]

    if ($unexpectedFieldsByPacket.ContainsKey($PacketName)) {
        foreach ($field in $unexpectedFieldsByPacket[$PacketName]) {
            if ($Object.PSObject.Properties.Name.Contains($field)) {
                Add-Finding -Category 'unexpected_field' -File $File -Detail ("unexpected field: {0}" -f $field)
            }
        }
    }
}

$jsonFiles = @()
foreach ($root in $Roots) {
    if (-not (Test-Path -LiteralPath $root)) {
        continue
    }

    $jsonFiles += Get-ChildItem -LiteralPath $root -Recurse -Filter '*.json' -File
}

foreach ($file in $jsonFiles) {
    if (-not (Test-JsonRootContextBoard -File $file.FullName)) {
        continue
    }

    $parsed = ConvertFrom-JsonObject -File $file.FullName
    if ($null -eq $parsed) {
        continue
    }

    if ($file.Name -like '*.validation.json') {
        Test-ValidationArtifact -File $file.FullName -Object $parsed
        continue
    }

    if (-not $packetSchemas.Contains($file.Name)) {
        Add-Finding -Category 'unknown_packet_name' -File $file.FullName -Detail ("unknown packet filename: {0}" -f $file.Name)
        continue
    }

    Test-PacketArtifact -File $file.FullName -Object $parsed -PacketName $file.Name
}

$sortedFindings = $findings | Sort-Object Category, File, Detail

Write-Host 'Packet contract validation'
Write-Host ("- checked_roots: {0}" -f $Roots.Count)
Write-Host ("- checked_json: {0}" -f $jsonFiles.Count)

if ($sortedFindings.Count -eq 0) {
    Write-Host '- findings: 0'
    exit 0
}

Write-Host ("- findings: {0}" -f $sortedFindings.Count)
foreach ($finding in $sortedFindings) {
    Write-Host ("[{0}] {1} - {2}" -f $finding.Category, $finding.File, $finding.Detail)
}

exit 1
