param(
    [Parameter(Mandatory = $true)]
    [string]$ChangeId,

    [string]$SessionName = ""
)

$ErrorActionPreference = "Stop"

. (Join-Path $PSScriptRoot "debugger-logger.ps1")

$session = New-DebuggerLoggerSession -ChangeId $ChangeId -SessionName $SessionName

Write-Output ("Created debugger logger session: " + $session.SessionDir)
Write-Output ("Events: " + $session.EventsPath)
Write-Output ("Notes: " + $session.NotesPath)
Write-Output ("Meta: " + $session.MetaPath)
