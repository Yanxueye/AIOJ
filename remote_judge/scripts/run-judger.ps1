$ErrorActionPreference = "Stop"

$env:REMOTE_JUDGE_GRPC_ADDR = "127.0.0.1:9090"

& "C:\Program Files\Go\bin\go.exe" run .\cmd\judger
