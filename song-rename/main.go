package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/raffleberry/tags"
	"github.com/raffleberry/tags/tag"
)

func main() {
	log.SetFlags(0)

	dryRun := flag.Bool("n", false, "print what would be renamed without renaming")
	flag.BoolVar(dryRun, "dry-run", false, "print what would be renamed without renaming")
	recursive := flag.Bool("r", false, "recurse into directories")
	flag.BoolVar(recursive, "recursive", false, "recurse into directories")
	quiet := flag.Bool("q", false, "only print errors")
	flag.BoolVar(quiet, "quiet", false, "only print errors")
	flag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: %s [-n] [-r] [-q] <file|dir>...\n\n", os.Args[0])
		fmt.Fprintf(os.Stderr, "Rename audio files to 'artist - title' (artists joined with \", \").\n\n")
		fmt.Fprintf(os.Stderr, "Flags:\n")
		flag.PrintDefaults()
	}
	flag.Parse()

	paths := flag.Args()
	if len(paths) == 0 {
		flag.Usage()
		os.Exit(2)
	}

	files, err := collectFiles(paths, *recursive)
	if err != nil {
		log.Fatal(err)
	}
	if len(files) == 0 {
		log.Fatal("no audio files found")
	}

	failed := 0
	for _, path := range files {
		newPath, unchanged, err := targetPath(path)
		if err != nil {
			log.Printf("%s: %v", path, err)
			failed++
			continue
		}
		if unchanged {
			if !*quiet {
				fmt.Printf("ok: %s\n", path)
			}
			continue
		}
		if *dryRun {
			fmt.Printf("%s -> %s\n", path, newPath)
			continue
		}
		if err := os.Rename(path, newPath); err != nil {
			log.Printf("%s: rename: %v", path, err)
			failed++
			continue
		}
		if !*quiet {
			fmt.Printf("%s -> %s\n", path, newPath)
		}
	}

	if failed > 0 {
		os.Exit(1)
	}
}

func collectFiles(paths []string, recursive bool) ([]string, error) {
	var files []string
	for _, p := range paths {
		info, err := os.Stat(p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		if !info.IsDir() {
			files = append(files, p)
			continue
		}
		if recursive {
			err := filepath.WalkDir(p, func(path string, d os.DirEntry, err error) error {
				if err != nil {
					return err
				}
				if !d.IsDir() && tags.LooksLike(path) {
					files = append(files, path)
				}
				return nil
			})
			if err != nil {
				return nil, err
			}
			continue
		}
		entries, err := os.ReadDir(p)
		if err != nil {
			return nil, fmt.Errorf("%s: %w", p, err)
		}
		for _, e := range entries {
			if e.IsDir() {
				continue
			}
			full := filepath.Join(p, e.Name())
			if tags.LooksLike(full) {
				files = append(files, full)
			}
		}
	}
	return files, nil
}

func targetPath(path string) (string, bool, error) {
	f, err := tags.Open(path)
	if err != nil {
		return "", false, err
	}

	artists := cleanValues(f.Tags().Values(tag.Artist))
	if len(artists) == 0 {
		artists = cleanValues(f.Tags().Values(tag.AlbumArtist))
	}
	title := strings.TrimSpace(f.Tags().Value(tag.Title))

	switch {
	case len(artists) == 0 && title == "":
		return "", false, fmt.Errorf("missing artist and title tags")
	case len(artists) == 0:
		return "", false, fmt.Errorf("missing artist tag")
	case title == "":
		return "", false, fmt.Errorf("missing title tag")
	}

	base := strings.Join(artists, ", ") + " - " + title
	base = sanitize(base)
	if base == "" || base == "-" || strings.Trim(base, " -.") == "" {
		return "", false, fmt.Errorf("artist/title produce an empty file name")
	}

	ext := filepath.Ext(path)
	newName := base + ext
	dir := filepath.Dir(path)

	if filepath.Base(path) == newName {
		return path, true, nil
	}

	dest := filepath.Join(dir, newName)
	return uniquePath(dest), false, nil
}

func cleanValues(values []string) []string {
	out := make([]string, 0, len(values))
	for _, v := range values {
		if v = strings.TrimSpace(v); v != "" {
			out = append(out, v)
		}
	}
	return out
}

func uniquePath(path string) string {
	if _, err := os.Stat(path); os.IsNotExist(err) {
		return path
	}
	ext := filepath.Ext(path)
	stem := strings.TrimSuffix(path, ext)
	for n := 1; ; n++ {
		candidate := fmt.Sprintf("%s (%d)%s", stem, n, ext)
		if _, err := os.Stat(candidate); os.IsNotExist(err) {
			return candidate
		}
	}
}

func sanitize(name string) string {
	name = strings.TrimSpace(name)
	var b strings.Builder
	for _, r := range name {
		switch {
		case r < 0x20:
			b.WriteRune('_')
		case strings.ContainsRune(`/\:*?"<>|`, r):
			b.WriteRune('_')
		default:
			b.WriteRune(r)
		}
	}
	out := strings.TrimSpace(b.String())
	// Windows dislikes trailing dots and spaces.
	out = strings.TrimRight(out, " .")
	// Collapse runs of underscores from replaced separators.
	for strings.Contains(out, "__") {
		out = strings.ReplaceAll(out, "__", "_")
	}
	// Keep names within a safe length for most filesystems.
	const maxRunes = 200
	if n := len([]rune(out)); n > maxRunes {
		out = string([]rune(out)[:maxRunes])
		out = strings.TrimSpace(strings.TrimRight(out, " ."))
	}
	return out
}
