$ErrorActionPreference = "Stop"

$env:REMOTE_JUDGE_REPOSITORY = "memory"
$env:REMOTE_JUDGE_QUEUE = "memory"
$env:REMOTE_JUDGE_JUDGER_MODE = "embedded"

& "C:\Program Files\Go\bin\go.exe" run .\cmd\server
