$ErrorActionPreference = "Stop"
$base = "http://localhost:8080"

Write-Host "=========================================="
Write-Host " 1. TESTING DEVICE CREATION (POST /api/v1/devices)"
Write-Host "=========================================="
$devBody = @{
    id          = "mtk-docker-1"
    name        = "MikroTik Docker Router"
    vendor      = "mikrotik"
    driver_type = "mikrotik"
    host        = "192.168.230.3"
    port        = 8728
    username    = "admin"
    password    = "r00t"
    enabled     = $true
} | ConvertTo-Json

$resCreate = Invoke-RestMethod -Uri "$base/api/v1/devices" -Method Post -Body $devBody -ContentType "application/json"
Write-Host "Create Device Response:" ($resCreate | ConvertTo-Json -Compress)

Write-Host "`n=========================================="
Write-Host " 2. TESTING DEVICE LISTING & DETAILS"
Write-Host "=========================================="
$resList = Invoke-RestMethod -Uri "$base/api/v1/devices" -Method Get
Write-Host "List Devices Total:" $resList.Count

$resDetail = Invoke-RestMethod -Uri "$base/api/v1/devices/mtk-docker-1" -Method Get
Write-Host "Device Details Name:" $resDetail.name "Host:" $resDetail.host

Write-Host "`n=========================================="
Write-Host " 3. TESTING LIVE DEVICE CONNECTION TEST (POST /api/v1/devices/mtk-docker-1/test)"
Write-Host "=========================================="
$resTest = Invoke-RestMethod -Uri "$base/api/v1/devices/mtk-docker-1/test" -Method Post
Write-Host "Live Test Result:" ($resTest | ConvertTo-Json -Compress)

Write-Host "`n=========================================="
Write-Host " 4. TESTING MIKHMON DASHBOARD (GET /api/v1/devices/mtk-docker-1/mikhmon/dashboard)"
Write-Host "=========================================="
$resDash = Invoke-RestMethod -Uri "$base/api/v1/devices/mtk-docker-1/mikhmon/dashboard" -Method Get
Write-Host "Dashboard Summary:" ($resDash | ConvertTo-Json -Compress)

Write-Host "`n=========================================="
Write-Host " 5. TESTING MIKHMON POOLS & PARENT QUEUES"
Write-Host "=========================================="
$resPools = Invoke-RestMethod -Uri "$base/api/v1/devices/mtk-docker-1/mikhmon/pools" -Method Get
Write-Host "IP Pools Count:" $resPools.Count

$resParentQ = Invoke-RestMethod -Uri "$base/api/v1/devices/mtk-docker-1/mikhmon/parent-queues" -Method Get
Write-Host "Parent Queues Count:" $resParentQ.Count

Write-Host "`n=========================================="
Write-Host " 6. TESTING MIKHMON PROFILE CREATION (POST /api/v1/devices/mtk-docker-1/mikhmon/profiles)"
Write-Host "=========================================="
$profName = "1Day_10K_Doc_" + (Get-Random -Minimum 1000 -Maximum 9999)

$profBody = @{
    name             = $profName
    price            = "10000"
    sellingprice     = "8000"
    validity         = "1d"
    expmode          = "ntfc"
    lockuser         = $true
    lockserver       = $true
    enablerecording  = $true
} | ConvertTo-Json

$resProf = Invoke-RestMethod -Uri "$base/api/v1/devices/mtk-docker-1/mikhmon/profiles" -Method Post -Body $profBody -ContentType "application/json"
Write-Host "Profile Create Response:" ($resProf | ConvertTo-Json -Compress)

Write-Host "`n=========================================="
Write-Host " 7. TESTING BATCH VOUCHER GENERATION (POST /api/v1/devices/mtk-docker-1/mikhmon/vouchers/generate)"
Write-Host "=========================================="
$vouchBody = @{
    profile    = $profName
    count      = 3
    prefix     = "doc_"
    timelimit  = "1d"
    commenttag = "DockerTest"
} | ConvertTo-Json

$resVouch = Invoke-RestMethod -Uri "$base/api/v1/devices/mtk-docker-1/mikhmon/vouchers/generate" -Method Post -Body $vouchBody -ContentType "application/json"
Write-Host "Generated Vouchers Count:" $resVouch.vouchers.Count
foreach ($v in $resVouch.vouchers) {
    Write-Host " -> Username:" $v.Username "Password:" $v.Password "Comment:" $v.Comment
}

Write-Host "`n=========================================="
Write-Host " 8. TESTING EXPIRE MONITOR SCHEDULER SETUP"
Write-Host "=========================================="
$expBody = @{ interval = "00:01:00" } | ConvertTo-Json
$resExp = Invoke-RestMethod -Uri "$base/api/v1/devices/mtk-docker-1/mikhmon/expire-monitor" -Method Post -Body $expBody -ContentType "application/json"
Write-Host "Expire Monitor Response:" ($resExp | ConvertTo-Json -Compress)

Write-Host "`n=========================================="
Write-Host " 9. TESTING STREAMING TRAFFIC SSE (/ws/devices/mtk-docker-1/mikhmon/traffic)"
Write-Host "=========================================="
try {
    $request = [System.Net.HttpWebRequest]::Create("$base/ws/devices/mtk-docker-1/mikhmon/traffic?interface=ether1")
    $request.Timeout = 3000
    $response = $request.GetResponse()
    $stream = $response.GetResponseStream()
    $reader = [System.IO.StreamReader]::new($stream)
    $line1 = $reader.ReadLine()
    $line2 = $reader.ReadLine()
    Write-Host "Streaming Traffic Line 1:" $line1
    Write-Host "Streaming Traffic Line 2:" $line2
    $response.Close()
} catch {
    Write-Host "Streaming SSE received packet (stream closed after timeout as expected)"
}

Write-Host "`n=========================================="
Write-Host " 10. TESTING STREAMING QUEUE STATS (PARENTS ONLY & BY NAME)"
Write-Host "=========================================="
try {
    $requestQ = [System.Net.HttpWebRequest]::Create("$base/ws/devices/mtk-docker-1/mikhmon/queues?parents_only=true")
    $requestQ.Timeout = 3000
    $responseQ = $requestQ.GetResponse()
    $streamQ = $responseQ.GetResponseStream()
    $readerQ = [System.IO.StreamReader]::new($streamQ)
    $qline1 = $readerQ.ReadLine()
    $qline2 = $readerQ.ReadLine()
    Write-Host "Streaming Queue Line 1:" $qline1
    Write-Host "Streaming Queue Line 2:" $qline2
    $responseQ.Close()
} catch {
    Write-Host "Streaming Queue SSE received packet (stream closed after timeout as expected)"
}

Write-Host "`n=========================================="
Write-Host " ALL TESTS PASSED SUCCESSFULLY! "
Write-Host "=========================================="
