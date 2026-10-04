# song-rename

Rename audio files to `artist - title`. Multiple artists become `a, b - title`.

Uses [github.com/raffleberry/tags](https://github.com/raffleberry/tags) for metadata (mp3, m4a/mp4, flac). Extension is preserved.

```sh
go run . song.mp3
song-rename *.flac ./album/
song-rename -n -r ./music/   # preview recursive run
```

Flags: `-n/--dry-run` preview, `-r/--recursive` walk directories, `-q/--quiet` errors only.

Rules:

- `artist` values joined with `", "`. Falls back to `albumartist` when `artist` is missing.
- Missing artist or title → file is skipped with an error.
- `/\:*?"<>|` and control chars → `_`, trailing dots/spaces trimmed, length capped.
- Never overwrites: `name (1).ext`, `name (2).ext`, …
- Already correct names print `ok:` and are left alone.
