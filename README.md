# store-core

Shared Go library backing [`store`](https://github.com/cushycush/store) (dotfile
symlinker) and [`stock`](https://github.com/cushycush/stock) (package
installer). Both tools read the same `.store/` directory, honour the same
`when:` platform filters, and invoke hooks with the same `STORE_*`
environment contract — this module is where those pieces live.

Not designed as a general-purpose library; the audience is the two binaries
above. The API is stable enough for internal use but may change without a
deprecation cycle if the consumers need it to.

## Packages

| Package | What it does |
|---|---|
| [`platform`](./platform) | Detect OS, arch, Linux distro (+ version), hostname, shell, WSL. Exposes the `STORE_*` env-var contract. |
| [`config`](./config) | `.store` directory layout, `FindRoot`, `ExpandHome`, and `WhenClause` (scalar-or-list YAML `when:` matching). |
| [`hooks`](./hooks) | Runs `<root>/.store/hooks/<name>` global scripts; exposes the `Env` helper so tool-specific hook runners can reuse the env contract. |
| [`ui`](./ui) | ANSI styling for CLI output — colors, bold/dim, doctor chips, prompts. Auto-disables on non-terminal stdout and honors `NO_COLOR` / `FORCE_COLOR`. |

## Example

```go
import (
    "fmt"

    "github.com/cushycush/store-core/config"
    "github.com/cushycush/store-core/platform"
)

func main() {
    root, _ := config.FindRoot(".")       // walk up until .store/ is found
    info := platform.Detect()             // linux, amd64, arch, …

    when := &config.WhenClause{OS: config.Strings{"linux", "darwin"}}
    fmt.Println(when.Matches(info))       // true on Linux or macOS
}
```

## `when:` matching

`when:` clauses gate a config entry on the detected platform. Every string
field accepts either a YAML scalar or a sequence. All specified fields must
match (AND); within a list-valued field, any entry matches (OR). Matching is
case-insensitive for string values.

```yaml
when:
  os: linux                  # scalar
  distro: [arch, ubuntu]     # sequence
  wsl: false
```

Fields: `os`, `arch`, `distro`, `distro_version`, `hostname`, `shell`, `wsl`.

## Hook env contract

Hook scripts under `.store/hooks/` receive:

| Variable | Value |
|---|---|
| `STORE_ROOT` | absolute path of the repo root containing `.store/` |
| `STORE_ACTION` | the action name passed by the caller (e.g. `link`, `install`) |
| `STORE_OS` | `runtime.GOOS` — `linux`, `darwin`, `windows` |
| `STORE_ARCH` | `runtime.GOARCH` — `amd64`, `arm64`, … |
| `STORE_DISTRO` | `/etc/os-release` ID on Linux; `macos` / `windows` otherwise |
| `STORE_DISTRO_VERSION` | distro version string where available |
| `STORE_HOSTNAME` | `os.Hostname()` |
| `STORE_SHELL` | basename of `$SHELL` (falls back to PowerShell / cmd detection on Windows) |
| `STORE_WSL` | `true` / `false` |

Tool-specific runners can layer extra variables on top via
`hooks.Env(root, action, "STORE_EXTRA=...")`.

## License

MIT. See [`store`'s LICENSE](https://github.com/cushycush/store/blob/main/LICENSE).
