[CmdletBinding()]
param()

$ErrorActionPreference = 'Stop'

$scriptRoot = Split-Path -Parent $PSCommandPath
$validator = Join-Path $scriptRoot 'validate-packet-contracts.ps1'
$fixturesRoot = Join-Path $scriptRoot 'fixtures\validate-packet-contracts'
$validRoot = Join-Path $fixturesRoot 'valid'
$invalidRoot = Join-Path $fixturesRoot 'invalid'

function Invoke-ValidatorProcess {
    param(
        [Parameter(Mandatory = $true)]
        [string]$Root
    )

    $output = & powershell -ExecutionPolicy Bypass -File $validator -Roots $Root 2>&1
    return [pscustomobject]@{
        ExitCode = $LASTEXITCODE
        Output = @($output)
    }
}

$validResult = Invoke-ValidatorProcess -Root $validRoot
if ($validResult.ExitCode -ne 0) {
    throw ("valid fixture failed validation:`n{0}" -f ($validResult.Output -join [Environment]::NewLine))
}

$invalidValidationFile = Join-Path $invalidRoot 'changes\invalid-change\context_board\impl-distill.packet.validation.json'
if (-not (Test-Path -LiteralPath $invalidValidationFile)) {
    throw 'invalid fixture is missing impl-distill.packet.validation.json'
}

$invalidValidation = Get-Content -LiteralPath $invalidValidationFile -Raw | ConvertFrom-Json
if ($invalidValidation.valid -ne $false) {
    throw 'invalid fixture validation artifact must set valid=false'
}

$invalidResult = Invoke-ValidatorProcess -Root $invalidRoot
if ($invalidResult.ExitCode -eq 0) {
    throw 'invalid fixture unexpectedly passed validation'
}

$joinedOutput = $invalidResult.Output -join [Environment]::NewLine
if ($joinedOutput -notmatch 'missing_required_field') {
    throw ("invalid fixture did not report missing_required_field:`n{0}" -f $joinedOutput)
}

Write-Host 'validate-packet-contracts tests: passed'
