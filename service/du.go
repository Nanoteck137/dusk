package service

import (
	"fmt"
	"io/fs"
	"path/filepath"
	"sort"
	"strings"
)

// DiskUsage returns the total size in bytes of the regular files under path.
// Unlike GNU du it does not count directory entries or special files, so the
// result is the sum of file sizes and is independent of the filesystem.
func DiskUsage(path string) (int64, error) {
	usage, err := ExtensionUsage(path)
	if err != nil {
		return 0, err
	}

	var total int64
	for _, size := range usage {
		total += size
	}

	return total, nil
}

// ExtensionUsage returns the total size in bytes of the regular files under
// path, grouped by file extension (including the leading dot). Hidden files
// and files without an extension are grouped under the empty string.
func ExtensionUsage(path string) (map[string]int64, error) {
	usage := make(map[string]int64)

	err := filepath.WalkDir(path, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		usage[ext(name)] += info.Size()
		return nil
	})

	return usage, err
}

// ext returns the file extension of name in lowercase, including the leading
// dot. Hidden files have no extension, and aliases are mapped to a canonical
// extension.
func ext(name string) string {
	base := filepath.Base(name)
	if strings.HasPrefix(base, ".") {
		return ""
	}

	e := strings.ToLower(filepath.Ext(name))
	if alias, ok := extAliases[e]; ok {
		return alias
	}

	return e
}

var extAliases = map[string]string{
	".jpeg": ".jpg",
}

// File is a regular file and its size in bytes.
type File struct {
	Path string
	Size int64
}

// LargestFiles returns the n largest regular files under path in descending
// order of size.
func LargestFiles(path string, n int) ([]File, error) {
	files := make([]File, 0)

	err := filepath.WalkDir(path, func(name string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}

		if !entry.Type().IsRegular() {
			return nil
		}

		info, err := entry.Info()
		if err != nil {
			return err
		}

		files = append(files, File{Path: name, Size: info.Size()})
		return nil
	})

	if err != nil {
		return nil, err
	}

	sort.Slice(files, func(i, j int) bool {
		if files[i].Size != files[j].Size {
			return files[i].Size > files[j].Size
		}

		return files[i].Path < files[j].Path
	})

	if n < len(files) {
		files = files[:n]
	}

	return files, nil
}

// HumanSize formats a size in bytes as a human readable string, e.g. 12K.
func HumanSize(bytes int64) string {
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%dB", bytes)
	}

	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}

	return fmt.Sprintf("%.1f%c", float64(bytes)/float64(div), "KMGTPE"[exp])
}
