$ErrorActionPreference = "Stop"

$env:REMOTE_JUDGE_REPOSITORY = "mysql"
$env:REMOTE_JUDGE_QUEUE = "rabbitmq"
$env:REMOTE_JUDGE_JUDGER_MODE = "embedded"
$env:REMOTE_JUDGE_MYSQL_DSN = "root:root@tcp(127.0.0.1:3306)/remote_judge?parseTime=true&charset=utf8mb4"
$env:REMOTE_JUDGE_RABBITMQ_URL = "amqp://guest:guest@127.0.0.1:5672/"

& "C:\Program Files\Go\bin\go.exe" run .\cmd\server
