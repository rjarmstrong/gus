# AGENTS.md

## Cursor Cloud specific instructions

`gus` is a **Go library** (package `gus`, module `github.com/rjarmstrong/gus`) for user
authentication and account management. There is no `main`, HTTP server, or UI — it is
embedded by other Go apps. "Running the app" means running the test suite and/or a small
consumer program against a database. See `README.md` for the public API overview.

### Toolchain / modules
- Built and tested with the base Go 1.22 toolchain. `go.mod`/`go.sum` are committed.
- Dependencies are intentionally **pinned to versions compatible with Go 1.22** and with the
  library's behavior:
  - `golang.org/x/crypto v0.4.0` — later versions make `bcrypt.GenerateFromPassword` **return an
    error for passwords > 72 bytes**. `SignUp` generates a 128-byte password, so newer `x/crypto`
    breaks `TestUsers_SignUp`. Do not bump `x/crypto` without also changing the (out-of-scope)
    library code.
  - `github.com/go-sql-driver/mysql v1.8.1` and `filippo.io/edwards25519 v1.1.0` — newer versions
    require Go >= 1.24.
- Because of the pins, plain `go vet/build/test ./...` run on Go 1.22 with **no** `GOTOOLCHAIN`
  override. Avoid `go mod tidy` / `go get -u`, which will re-resolve to newer, incompatible
  versions and pull a Go 1.25 toolchain.

### Database (required for tests)
- Tests (`main_test.go`) connect over TCP to **MySQL on 127.0.0.1:3306**, DB `gus_test`,
  user `root`, password `rootPassword`. MariaDB 10.11 is installed as a drop-in replacement and
  its data dir (root password + `gus_test`/`gus` databases) is part of the VM snapshot.
- The server is **not auto-started**. Start it each session before running tests:
  ```bash
  sudo service mariadb start
  ```
- The `Seed: true` option drops & recreates all tables on every `GetDb`, so test/demo runs are
  self-resetting.

### Lint / build / test
```bash
go vet ./...      # clean except pre-existing db.go:138 "non-constant format string" (harmless on 1.22)
go build ./...
go test ./...     # requires MariaDB running (see above)
```

### Quick end-to-end sanity check (consumer flow)
A consumer must blank-import the driver (`_ "github.com/go-sql-driver/mysql"`) since the library
does not. Minimal program: `GetDb(mysql, Seed:true)` -> `NewOrgs`/`NewUsers` -> `Create` org ->
`SignUp` -> `SignIn`. This exercises the core auth path against MySQL.
