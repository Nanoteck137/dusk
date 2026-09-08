package service

import (
	"fmt"
	"io/fs"
	"path/filepath"
)

// DiskUsage returns the total size in bytes of the regular files under path.
// Unlike GNU du it does not count directory entries or special files, so the
// result is the sum of file sizes and is independent of the filesystem.
func DiskUsage(path string) (int64, error) {
	var total int64

	err := filepath.WalkDir(path, func(_ string, entry fs.DirEntry, err error) error {
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

		total += info.Size()
		return nil
	})

	return total, err
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
