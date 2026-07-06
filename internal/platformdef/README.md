# Platform Definitions

Custom scrapligo YAML platform definitions for vendors without built-in
support. See `TECH-STACK-DAN-PERSIAPAN.md` §7 before adding a new file here.

Naming: `<vendor>_<model>.yaml` (e.g. `zte_c320.yaml`).

Note: this folder holds scrapligo *prompt-pattern* definitions only. Command
catalog/translation and risk classification for each vendor live in that
vendor's own `internal/driver/<vendor>/commands.go`, not here.
