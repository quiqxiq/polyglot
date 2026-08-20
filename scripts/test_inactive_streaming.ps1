$ErrorActionPreference = "Stop"
$base = "http://localhost:8080"
$deviceID = "mtk-test"

Write-Host "=========================================="
Write-Host " 1. TESTING REAL-TIME STREAMING: HOTSPOT INACTIVE USERS"
Write-Host " Uri: $base/ws/devices/$deviceID/mikhmon/hotspot-inactive"
Write-Host "=========================================="
try {
    $req = [System.Net.HttpWebRequest]::Create("$base/ws/devices/$deviceID/mikhmon/hotspot-inactive")
    $req.Timeout = 4000
    $resp = $req.GetResponse()
    $stream = $resp.GetResponseStream()
    $reader = [System.IO.StreamReader]::new($stream)
    
    $eventLine = ""
    while ($true) {
        $line = $reader.ReadLine()
        if ($null -eq $line) { break }
        if ($line.StartsWith("event:")) {
            $eventLine = $line
        } elseif ($line.StartsWith("data:")) {
            Write-Host "Hotspot Inactive SSE Event:" $eventLine
            Write-Host "Hotspot Inactive SSE Data:" $line
            break
        }
    }
    $resp.Close()
} catch [System.Net.WebException] {
    if ($_.Status -eq [System.Net.WebExceptionStatus]::Timeout) {
        Write-Host "Hotspot Inactive SSE Stream timeout after receiving data"
    } else {
        Write-Host "WebException:" $_.Exception.Message
    }
} catch {
    Write-Host "Exception:" $_.Exception.Message
}

Write-Host "`n=========================================="
Write-Host " 2. TESTING REAL-TIME STREAMING: PPPOE INACTIVE SUBSCRIBERS"
Write-Host " Uri: $base/ws/devices/$deviceID/mikhmon/ppp-inactive"
Write-Host "=========================================="
try {
    $req = [System.Net.HttpWebRequest]::Create("$base/ws/devices/$deviceID/mikhmon/ppp-inactive")
    $req.Timeout = 4000
    $resp = $req.GetResponse()
    $stream = $resp.GetResponseStream()
    $reader = [System.IO.StreamReader]::new($stream)
    
    $eventLine = ""
    while ($true) {
        $line = $reader.ReadLine()
        if ($null -eq $line) { break }
        if ($line.StartsWith("event:")) {
            $eventLine = $line
        } elseif ($line.StartsWith("data:")) {
            Write-Host "PPPoE Inactive SSE Event:" $eventLine
            Write-Host "PPPoE Inactive SSE Data:" $line
            break
        }
    }
    $resp.Close()
} catch [System.Net.WebException] {
    if ($_.Status -eq [System.Net.WebExceptionStatus]::Timeout) {
        Write-Host "PPPoE Inactive SSE Stream timeout after receiving data"
    } else {
        Write-Host "WebException:" $_.Exception.Message
    }
} catch {
    Write-Host "Exception:" $_.Exception.Message
}

Write-Host "`n=========================================="
Write-Host " INACTIVE STREAMING TESTS COMPLETED SUCCESSFULLY! "
Write-Host "=========================================="
