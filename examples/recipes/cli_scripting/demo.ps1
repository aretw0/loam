# Loam CLI Scripting Demo (PowerShell)
# Prerequisites: loam

$VaultDir = "demo_vault_ps"
if (Test-Path $VaultDir) { Remove-Item -Recurse -Force $VaultDir }
New-Item -ItemType Directory -Force -Path $VaultDir | Out-Null

Write-Host "--- 1. Initializing Vault ---" -ForegroundColor Cyan
loam init $VaultDir --nover
Push-Location $VaultDir

Write-Host "`n--- 2. Simple Piping (Text) ---" -ForegroundColor Cyan
"Build started at $(Get-Date)" | loam write --id logs/build-1 --message "start build"
loam read logs/build-1

Write-Host "`n--- 3. JSON Pipeline (PowerShell Object -> loam) ---" -ForegroundColor Cyan
$pkg = @{ package = "loam"; version = "0.8.5"; status = "stable" }
# Convert to JSON and save
$pkg | ConvertTo-Json | loam write --id meta/pkg.json --raw
Write-Host "Saved meta/pkg.json:"
loam read meta/pkg.json

Write-Host "`n--- 4. CSV Splitting (Loop) ---" -ForegroundColor Cyan
# Create a dummy CSV
@"
id,title,priority
task-1,Fix Bugs,High
task-2,Write Docs,Medium
task-3,Release,Critical
"@ | Set-Content tasks.csv

# Import-Csv automatically parses headers
Import-Csv tasks.csv | ForEach-Object {
    $content = "# $($_.title)`n`nPriority: $($_.priority)"
    loam write --id "tasks/$($_.id).md" --content $content --message "import task $($_.id)"
    Write-Host "Imported $($_.id)"
}

loam list

Write-Host "`n--- Demo Complete ---" -ForegroundColor Cyan
Pop-Location
Remove-Item -Recurse -Force $VaultDir
