module examples/basic-app

go 1.25.4

// Use the local version of Loam by default for development/testing.
// To use the published version:
// 1. Comment out or remove the 'replace' line.
// 2. Run 'go get github.com/aretw0/loam@latest'
replace github.com/aretw0/loam => ../../../

require github.com/aretw0/loam v0.0.0-00010101000000-000000000000

require gopkg.in/yaml.v3 v3.0.1 // indirect
