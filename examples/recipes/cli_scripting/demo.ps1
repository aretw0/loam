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
    $content = "# $($_.title)`n`nPriority: $($_.priority)`n"
    loam write --id "tasks/$($_.id).md" --content $content --message "import task $($_.id)"
    Write-Host "Imported $($_.id)"
}

loam list

Write-Host "`n--- 5. Complex CSV (PowerShell Idioms) ---" -ForegroundColor Cyan
# Create a "dirty" CSV
@"
ID,Title,DESCRIPTION,Internal_Code
feat-1,Add Logging,Implement zap logger,SEC-001
feat-2,Refactor API,Clean up handlers,TECH-99
"@ | Set-Content dirty.csv

# Import, Transform, and Save
Import-Csv dirty.csv | 
Select-Object @{N = 'id'; E = { "features/$($_.ID)" } }, 
@{N = 'title'; E = { $_.Title } }, 
@{N = 'content'; E = { $_.DESCRIPTION } } | ForEach-Object {
    $id = $_.id
    $content = $_.content
        
    # Build YAML Frontmatter dynamically from other properties
    $yaml = "---`n"
    $_.PSObject.Properties | Where-Object { $_.Name -notin 'id', 'content' } | ForEach-Object {
        $yaml += "$($_.Name): $($_.Value)`n"
    }
    $yaml += "---`n`n"
        
    $doc = $yaml + $content
        
    # We explicitly use .md extension so loam parses the frontmatter we just built
    $doc | loam write --id "$id.md" --raw --message "import feature $id"
    Write-Host "Imported $id (Cleaned)"
}

Write-Host "`n--- 6. Complex CSV (Loam Set Flags) ---" -ForegroundColor Cyan
# Re-using dirty.csv, but cleaner approach using --set
Import-Csv dirty.csv | ForEach-Object {
    $id = "set_demo/$($_.ID)"
    # Clean up title (example)
    $title = $_.Title.Trim()
    
    # Use --set for metadata, --content for body
    # This avoids manual YAML construction!
    loam write --id "$id.md" --set "title=$title" --set "internal_code=$($_.Internal_Code)" --content "$($_.DESCRIPTION)`n" --message "import feature $id via flags"
    Write-Host "Imported $id (Flags)"
}

loam read set_demo/feat-1.md

Write-Host "`n--- Demo Complete ---"

loam read features/feat-1


Write-Host "`n--- Demo Complete ---" -ForegroundColor Cyan
Pop-Location
Remove-Item -Recurse -Force $VaultDir
