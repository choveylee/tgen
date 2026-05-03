# tgen

`tgen` is a Go-based scaffold generator for backend services built on the Choveylee service stack. It copies the local template, replaces repository placeholders, and initializes the generated module with the dependency pins required by the current scaffold.

## Features

- Copies the canonical scaffold from `template/` into the target service directory
- Replaces placeholders such as `{{domain}}`, `{{app_name}}`, `{{app_name2}}`, and `{{APP_NAME}}`
- Initializes the generated Go module and pins `gorm.io/plugin/dbresolver@v1.6.2` to preserve compatibility with the current GORM dependency set
- Relies on the built-in `/healthz` endpoint provided by `tserver` instead of generating a duplicate application route

## Requirements

- Go `1.25.0` or later

## Usage

The current workspace configuration generates the sample service in `test-backend/`:

```bash
env GOCACHE=/private/tmp/gocache GOMODCACHE=/private/tmp/gomodcache go run .
```

## Repository Layout

- `main.go` contains the scaffold generator entry point
- `template/` contains the canonical source template for generated services
- `test-backend/` contains the generated sample service used for local validation in this workspace

## Maintenance

- Treat `template/` as the source of truth for scaffold changes
- Regenerate or resynchronize `test-backend/` after template changes when rendered output needs to be validated
- Keep generated-service documentation and log messages aligned with the conventions used by the Choveylee dependency libraries

## License

This repository does not currently declare a license.
