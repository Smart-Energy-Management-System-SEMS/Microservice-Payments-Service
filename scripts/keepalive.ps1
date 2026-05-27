param(
    [string]$Url = $env:KEEPALIVE_URL,
    [int]$IntervalSeconds = $(if ($env:KEEPALIVE_INTERVAL_SECONDS) { [int]$env:KEEPALIVE_INTERVAL_SECONDS } else { 600 })
)

if ([string]::IsNullOrWhiteSpace($Url)) {
    Write-Error "KEEPALIVE_URL is required. Example: .\scripts\keepalive.ps1 -Url https://your-service.onrender.com/health"
    exit 1
}

Write-Host "Starting keep-alive pings to $Url every $IntervalSeconds seconds"

while ($true) {
    $timestamp = (Get-Date).ToUniversalTime().ToString("yyyy-MM-ddTHH:mm:ssZ")
    try {
        $response = Invoke-WebRequest -Uri $Url -Method GET -TimeoutSec 20 -UseBasicParsing
        Write-Host "[$timestamp] GET $Url -> HTTP $($response.StatusCode)"
    }
    catch {
        Write-Host "[$timestamp] GET $Url -> ERROR $($_.Exception.Message)"
    }
    Start-Sleep -Seconds $IntervalSeconds
}
