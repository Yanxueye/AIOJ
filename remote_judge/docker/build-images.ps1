$ErrorActionPreference = "Stop"

$root = Split-Path -Parent $MyInvocation.MyCommand.Path
$imagesRoot = Join-Path $root "images"

$targets = @(
    @{ Name = "remote-judge-cpp17"; Path = Join-Path $imagesRoot "cpp17" },
    @{ Name = "remote-judge-go122"; Path = Join-Path $imagesRoot "go1.22" },
    @{ Name = "remote-judge-python311"; Path = Join-Path $imagesRoot "python3.11" }
)

foreach ($target in $targets) {
    Write-Host "Building $($target.Name) from $($target.Path)"
    docker build -t $target.Name $target.Path
}

Write-Host "All remote_judge images built successfully."
