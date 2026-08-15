$files = Get-ChildItem -Path . -Filter *.go -Recurse | Where-Object { $_.FullName -notmatch '\\.git|\\.ref|\\.agents|vendor' }

Write-Host "=== Active Polyglot Project Go Files: $($files.Count) ==="

$over400 = 0
$over500 = 0
$maxLines = 0
$maxFile = ""

foreach ($f in $files) {
    $lines = (Get-Content $f.FullName | Measure-Object -Line).Lines
    if ($lines -gt $maxLines) {
        $maxLines = $lines
        $maxFile = $f.FullName.Replace((Get-Location).Path + "\", "")
    }
    if ($lines -gt 500) {
        $over500++
        Write-Host "[VIOLATION >500]: $lines lines in $($f.FullName)"
    } elseif ($lines -gt 400) {
        $over400++
        Write-Host "[WARNING >400]: $lines lines in $($f.FullName)"
    }
}

Write-Host "Maximum lines in active codebase: $maxLines lines ($maxFile)"
Write-Host "Active Files > 500 lines: $over500"
Write-Host "Active Files > 400 lines: $over400"
