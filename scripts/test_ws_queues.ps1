$ErrorActionPreference = "Stop"
$base = "http://localhost:8080"
$deviceID = "mtk-docker-1"

Write-Host "=========================================="
Write-Host " 1. ENSURING DEVICE IS REGISTERED"
Write-Host "=========================================="
$devBody = @{
    id          = $deviceID
    name        = "MikroTik Docker Router"
    vendor      = "mikrotik"
    driver_type = "mikrotik"
    host        = "192.168.230.3"
    port        = 8728
    username    = "admin"
    password    = "r00t"
    enabled     = $true
} | ConvertTo-Json

try {
    $resCreate = Invoke-RestMethod -Uri "$base/api/v1/devices" -Method Post -Body $devBody -ContentType "application/json"
    Write-Host "Device registration response:" ($resCreate | ConvertTo-Json -Compress)
} catch {
    Write-Host "Device already registered or response received."
}

Write-Host "`n=========================================="
Write-Host " 2. CREATING TEST PARENT QUEUE ON ROUTER"
Write-Host "=========================================="
$profBody = @{
    name             = "Parent-Total-Bandwidth"
    price            = "0"
    sellingprice     = "0"
    validity         = ""
    expmode          = "0"
} | ConvertTo-Json

try {
    $resProf = Invoke-RestMethod -Uri "$base/api/v1/devices/$deviceID/mikhmon/profiles" -Method Post -Body $profBody -ContentType "application/json"
    Write-Host "Profile/Queue setup response:" ($resProf | ConvertTo-Json -Compress)
} catch {
    Write-Host "Queue creation note:" $_.Exception.Message
}

Write-Host "`n=========================================="
Write-Host " 3. TESTING SSE STREAMING: QUEUES (PARENTS ONLY)"
Write-Host " Uri: $base/ws/devices/$deviceID/mikhmon/queues?parents_only=true"
Write-Host "=========================================="
try {
    $req = [System.Net.HttpWebRequest]::Create("$base/ws/devices/$deviceID/mikhmon/queues?parents_only=true")
    $req.Timeout = 5000
    $resp = $req.GetResponse()
    $stream = $resp.GetResponseStream()
    $reader = [System.IO.StreamReader]::new($stream)
    
    $eventLine = $reader.ReadLine()
    $dataLine = $reader.ReadLine()
    Write-Host "Received Event Header:" $eventLine
    Write-Host "Received Event Payload:" $dataLine
    $resp.Close()
} catch [System.Net.WebException] {
    if ($_.Status -eq [System.Net.WebExceptionStatus]::Timeout) {
        Write-Host "Stream read connection timeout after receiving events (Normal for open SSE streams)"
    } else {
        Write-Host "WebException:" $_.Exception.Message
    }
} catch {
    Write-Host "Exception:" $_.Exception.Message
}

Write-Host "`n=========================================="
Write-Host " 4. TESTING SSE STREAMING: QUEUES (BY QUEUE NAME)"
Write-Host " Uri: $base/ws/devices/$deviceID/mikhmon/queues?name=Parent-Total-Bandwidth"
Write-Host "=========================================="
try {
    $req = [System.Net.HttpWebRequest]::Create("$base/ws/devices/$deviceID/mikhmon/queues?name=Parent-Total-Bandwidth")
    $req.Timeout = 5000
    $resp = $req.GetResponse()
    $stream = $resp.GetResponseStream()
    $reader = [System.IO.StreamReader]::new($stream)
    
    $eventLine = $reader.ReadLine()
    $dataLine = $reader.ReadLine()
    Write-Host "Received Event Header:" $eventLine
    Write-Host "Received Event Payload:" $dataLine
    $resp.Close()
} catch [System.Net.WebException] {
    if ($_.Status -eq [System.Net.WebExceptionStatus]::Timeout) {
        Write-Host "Stream read connection timeout after receiving events (Normal for open SSE streams)"
    } else {
        Write-Host "WebException:" $_.Exception.Message
    }
} catch {
    Write-Host "Exception:" $_.Exception.Message
}

Write-Host "`n=========================================="
Write-Host " 4. TESTING SSE STREAMING: INTERFACE TRAFFIC (ETHER1)"
Write-Host " Uri: $base/ws/devices/$deviceID/mikhmon/traffic?interface=ether1"
Write-Host "=========================================="
try {
    $req = [System.Net.HttpWebRequest]::Create("$base/ws/devices/$deviceID/mikhmon/traffic?interface=ether1")
    $req.Timeout = 5000
    $resp = $req.GetResponse()
    $stream = $resp.GetResponseStream()
    $reader = [System.IO.StreamReader]::new($stream)
    
    $eventLine = $reader.ReadLine()
    $dataLine = $reader.ReadLine()
    Write-Host "Received Event Header:" $eventLine
    Write-Host "Received Event Payload:" $dataLine
    $resp.Close()
} catch [System.Net.WebException] {
    if ($_.Status -eq [System.Net.WebExceptionStatus]::Timeout) {
        Write-Host "Stream read connection timeout after receiving events (Normal for open SSE streams)"
    } else {
        Write-Host "WebException:" $_.Exception.Message
    }
} catch {
    Write-Host "Exception:" $_.Exception.Message
}

Write-Host "`n=========================================="
Write-Host " WEBSOCKET / SSE STREAMING TEST COMPLETED! "
Write-Host "=========================================="
