Set-StrictMode -Version Latest
$ErrorActionPreference = "Stop"

function New-DebuggerLoggerSession {
    param(
        [Parameter(Mandatory = $true)]
        [string]$ChangeId,

        [string]$SessionName = "",

        [string]$RootPath = ""
    )

    if ([string]::IsNullOrWhiteSpace($RootPath)) {
        $RootPath = Split-Path -Parent (Split-Path -Parent (Split-Path -Parent $PSScriptRoot))
    }

    if ([string]::IsNullOrWhiteSpace($SessionName)) {
        $SessionName = "session-" + (Get-Date).ToString("yyyyMMdd-HHmmss")
    }

    $changeDir = Join-Path $RootPath ("changes\" + $ChangeId)
    $debuggerRoot = Join-Path $changeDir "debugger_logs"
    $sessionDir = Join-Path $debuggerRoot $SessionName
    $eventsPath = Join-Path $sessionDir "events.jsonl"
    $notesPath = Join-Path $sessionDir "notes.md"
    $metaPath = Join-Path $sessionDir "session.json"

    New-Item -ItemType Directory -Path $debuggerRoot -Force | Out-Null
    New-Item -ItemType Directory -Path $sessionDir -Force | Out-Null

    $meta = [ordered]@{
        change_id = $ChangeId
        session_name = $SessionName
        created_at = (Get-Date).ToString("o")
        events_path = $eventsPath
        notes_path = $notesPath
    }

    if (-not (Test-Path $eventsPath)) {
        New-Item -ItemType File -Path $eventsPath -Force | Out-Null
    }

    if (-not (Test-Path $notesPath)) {
        @(
            "# Debugger Notes"
            ""
            "## Hypotheses"
            "- "
            ""
            "## Observations"
            "- "
            ""
            "## Next Check"
            "- "
        ) | Set-Content -Path $notesPath -Encoding utf8
    }

    $meta | ConvertTo-Json -Depth 4 | Set-Content -Path $metaPath -Encoding utf8

    [pscustomobject]@{
        ChangeId = $ChangeId
        SessionName = $SessionName
        SessionDir = $sessionDir
        EventsPath = $eventsPath
        NotesPath = $notesPath
        MetaPath = $metaPath
    }
}

function Write-DebuggerEvent {
    param(
        [Parameter(Mandatory = $true)]
        [string]$EventsPath,

        [Parameter(Mandatory = $true)]
        [string]$Stage,

        [Parameter(Mandatory = $true)]
        [string]$Message,

        [ValidateSet("trace", "debug", "info", "warn", "error")]
        [string]$Level = "info",

        [hashtable]$Fields = @{}
    )

    if (-not (Test-Path $EventsPath)) {
        throw "Events file not found: $EventsPath"
    }

    $event = [ordered]@{
        timestamp = (Get-Date).ToString("o")
        level = $Level
        stage = $Stage
        message = $Message
        fields = $Fields
    }

    Add-Content -Path $EventsPath -Value ($event | ConvertTo-Json -Compress -Depth 6) -Encoding utf8
}
