# tgen

`tgen` is a Go-based scaffold generator for backend services built on the Choveylee service stack. It copies the local template, replaces repository placeholders, and initializes the generated module with the dependency pins required by the current scaffold.

## Features

- Copies the canonical scaffold from `template/` into the target service directory
- Replaces placeholders such as `{{domain}}`, `{{app_name}}`, `{{app_name2}}`, and `{{APP_NAME}}`
- Initializes the generated Go module from the versions and checksums in `template/go.mod` and `template/go.sum`, including the GORM resolver pin
- Relies on the built-in `/healthz` endpoint provided by `tserver` instead of generating a duplicate application route

## Requirements

- Go `1.27.4` or later
- POSIX shell (`/bin/sh`)

## Usage

Set `domain` and `appName` in `main.go` before running. The current target is `knowledge-base-backend/`. Generate from the directory containing `template/`. The target must be absent or empty. The generator publishes the directory after rendering and module initialization succeed.

```bash
go run .
```

The generator uses the caller's `GOPROXY`. If a pinned Choveylee commit is available only locally, use a Go module proxy containing that revision. The generated Dockerfile accepts a `GOPROXY` build argument for the same purpose.

### Run the generated scaffold

Build the service with the executable name used by its config file. For the default project name, run these commands from the generated project directory.

```bash
go build -o build/knowledge_base_backend ./cmd
cp cmd/knowledge_base_backend_config.ini build/
./build/knowledge_base_backend
```

Fill in the generated MySQL and Redis settings before starting the service. On first startup, the migrations create the sample `template` table used by the ORM generator. To build the sample queries, configure `cmd/orm-gen/orm_gen_config.ini` for that database, copy it next to the executable, and run the command from the generated project directory.

```bash
go build -o build/orm-gen ./cmd/orm-gen
cp cmd/orm-gen/orm_gen_config.ini build/
./build/orm-gen
go mod tidy
```

Use `sh ./build.sh` to build and assemble the Linux amd64 release files.

## Repository Layout

- `main.go` contains the scaffold generator entry point
- `template/` contains the canonical source template for generated services

## Maintenance

- Treat `template/` as the source of truth for scaffold changes
- Render into a fresh temporary directory to validate template changes; migrate existing generated projects explicitly
- Keep generated-service documentation and log messages aligned with the conventions used by the Choveylee dependency libraries

## License

This repository does not currently declare a license.
