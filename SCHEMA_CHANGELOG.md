# JSON Schema Changelog

Both `diskwhy scan --json` and `diskwhy clean --json` output a versioned JSON object.
The `schema_version` field lets consumers detect breaking changes.

---

## schema_version: 1

Initial stable schema. All fields are present in every response.

### Scan output (`diskwhy scan --json`)

```json
{
  "schema_version": 1,
  "scanned_at": "2026-01-15T10:30:00Z",
  "scan_mode": "quick",
  "header": "[macOS / Macintosh HD]",
  "disk": {
    "total_bytes": 499963174912,
    "used_bytes":  312345678901,
    "free_bytes":  187617496011,
    "mount":       "/"
  },
  "items": [
    {
      "path":             "/Users/you/projects/app/node_modules",
      "category":         "node_modules",
      "size_bytes":       523456789,
      "staleness_score":  "unused",
      "staleness_source": "atime"
    }
  ],
  "docker": {
    "dangling_images":   3,
    "dangling_bytes":    1073741824,
    "unused_volumes":    2,
    "unused_vol_bytes":  524288000,
    "containers_total":  5,
    "containers_stopped": 2
  },
  "summary": {
    "total_items":      12,
    "total_bytes":      4831838208,
    "elapsed_ms":       842
  }
}
```

`docker` is `null` when Docker is not available or `SKIP_DOCKER=1` is set.

**`scan_mode`** values: `"quick"` | `"deep"` | `"path"`

**`staleness_score`** values: `"active"` | `"recent"` | `"stale"` | `"unused"` | `"unknown"`

**`staleness_source`** values: `"atime"` | `"mtime"` | `"none"`

**`category`** values:
`"node_modules"` | `"git_objects"` | `"docker"` | `"pycache"` | `"pip_cache"` |
`"npm_cache"` | `"brew_cache"` | `"xcode_derived"` | `"apt_cache"` | `"snap_cache"` |
`"logs"` | `"trash"`

---

### Clean output (`diskwhy clean --json`)

```json
{
  "schema_version": 1,
  "cleaned_at":         "2026-01-15T10:31:00Z",
  "dry_run":            false,
  "use_trash":          false,
  "results": [
    {
      "path":       "/Users/you/projects/app/node_modules",
      "category":   "node_modules",
      "outcome":    "deleted",
      "size_bytes": 523456789,
      "error":      ""
    }
  ],
  "docker_freed_bytes": 1073741824,
  "summary": {
    "deleted":      8,
    "trashed":      0,
    "skipped":      2,
    "errors":       0,
    "freed_bytes":  4831838208
  }
}
```

**`outcome`** values:
`"dry_run"` | `"skipped"` | `"trashed"` | `"deleted"` | `"gc_run"` | `"error"`

`"gc_run"` is used for `git_objects` items — `git gc` was run rather than deleting files directly.

`"skipped"` means the item matched no enabled category, was `active` and not a cache, or was on the blocklist.

`docker_freed_bytes` is `0` when Docker prune was not requested or Docker is unavailable.

---

## Versioning policy

- **Additive changes** (new optional fields, new enum values) do not increment `schema_version`.
- **Breaking changes** (removed fields, renamed fields, changed types) increment `schema_version`.
- Consumers should ignore unknown fields to remain forward-compatible.
