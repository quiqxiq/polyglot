$ErrorActionPreference = "Stop"
$base = "http://localhost:8080"
$deviceID = "mtk-test"

Write-Host "=========================================="
Write-Host " 1. TESTING SYSTEM RESOURCE STREAMING (interval=1s)"
Write-Host " Uri: $base/ws/devices/$deviceID/mikhmon/resource"
Write-Host "=========================================="
try {
    $req = [System.Net.HttpWebRequest]::Create("$base/ws/devices/$deviceID/mikhmon/resource")
    $req.Timeout = 4000
    $resp = $req.GetResponse()
    $stream = $resp.GetResponseStream()
    $reader = [System.IO.StreamReader]::new($stream)
    
    $eventLine = $reader.ReadLine()
    $dataLine = $reader.ReadLine()
    Write-Host "Resource SSE Event:" $eventLine
    Write-Host "Resource SSE Data:" $dataLine
    $resp.Close()
} catch [System.Net.WebException] {
    if ($_.Status -eq [System.Net.WebExceptionStatus]::Timeout) {
        Write-Host "Resource SSE Stream timeout after receiving data"
    } else {
        Write-Host "WebException:" $_.Exception.Message
    }
} catch {
    Write-Host "Exception:" $_.Exception.Message
}

Write-Host "`n=========================================="
Write-Host " 2. TESTING HOTSPOT USERS STREAMING (DIRECTORY)"
Write-Host " Uri: $base/ws/devices/$deviceID/mikhmon/hotspot-users"
Write-Host "=========================================="
try {
    $req = [System.Net.HttpWebRequest]::Create("$base/ws/devices/$deviceID/mikhmon/hotspot-users")
    $req.Timeout = 4000
    $resp = $req.GetResponse()
    $stream = $resp.GetResponseStream()
    $reader = [System.IO.StreamReader]::new($stream)
    
    $eventLine = $reader.ReadLine()
    $dataLine = $reader.ReadLine()
    Write-Host "Hotspot Users SSE Event:" $eventLine
    Write-Host "Hotspot Users SSE Data:" $dataLine
    $resp.Close()
} catch [System.Net.WebException] {
    if ($_.Status -eq [System.Net.WebExceptionStatus]::Timeout) {
        Write-Host "Hotspot Users SSE Stream timeout after receiving data"
    } else {
        Write-Host "WebException:" $_.Exception.Message
    }
} catch {
    Write-Host "Exception:" $_.Exception.Message
}

Write-Host "`n=========================================="
Write-Host " 3. TESTING PPPOE SECRETS STREAMING (DIRECTORY)"
Write-Host " Uri: $base/ws/devices/$deviceID/mikhmon/ppp-secrets"
Write-Host "=========================================="
try {
    $req = [System.Net.HttpWebRequest]::Create("$base/ws/devices/$deviceID/mikhmon/ppp-secrets")
    $req.Timeout = 4000
    $resp = $req.GetResponse()
    $stream = $resp.GetResponseStream()
    $reader = [System.IO.StreamReader]::new($stream)
    
    $eventLine = $reader.ReadLine()
    $dataLine = $reader.ReadLine()
    Write-Host "PPPoE Secrets SSE Event:" $eventLine
    Write-Host "PPPoE Secrets SSE Data:" $dataLine
    $resp.Close()
} catch [System.Net.WebException] {
    if ($_.Status -eq [System.Net.WebExceptionStatus]::Timeout) {
        Write-Host "PPPoE Secrets SSE Stream timeout after receiving data"
    } else {
        Write-Host "WebException:" $_.Exception.Message
    }
} catch {
    Write-Host "Exception:" $_.Exception.Message
}

Write-Host "`n=========================================="
Write-Host " 4. TESTING INCOME SUMMARY REST API"
Write-Host " Uri: $base/api/v1/devices/$deviceID/mikhmon/income"
Write-Host "=========================================="
try {
    $resIncome = Invoke-RestMethod -Uri "$base/api/v1/devices/$deviceID/mikhmon/income" -Method Get
    Write-Host "Income Response:" ($resIncome | ConvertTo-Json -Compress)
} catch {
    Write-Host "Income REST Exception:" $_.Exception.Message
}

Write-Host "`n=========================================="
Write-Host " ALL PURE STREAMING TESTS COMPLETED! "
Write-Host "=========================================="
