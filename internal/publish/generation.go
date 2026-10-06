package publish

import (
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
)

// ResolveGeneration pins all publication reads to the same physical directory.
func ResolveGeneration(path string) (string, error) {
	resolved, err := filepath.EvalSymlinks(path)
	if err != nil {
		return "", err
	}
	return filepath.Abs(resolved)
}

// Activate swaps only a symlink. A real directory is never moved out of the way.
// On Linux rename is atomic: readers see either the old or the new generation.
func Activate(current, generation string) error {
	generation, err := ResolveGeneration(generation)
	if err != nil {
		return err
	}
	if _, err := os.Stat(filepath.Join(generation, "tokens.sqlite")); err != nil {
		return err
	}
	if _, err := Load(filepath.Join(generation, "tokens.sqlite")); err != nil {
		return fmt.Errorf("validate generation: %w", err)
	}
	if err := syncGeneration(generation); err != nil {
		return fmt.Errorf("sync generation before activation: %w", err)
	}
	if info, err := os.Lstat(current); err == nil && info.Mode()&os.ModeSymlink == 0 {
		return errors.New("current publication must be a symlink, not an existing file or directory")
	} else if err != nil && !os.IsNotExist(err) {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(current), 0700); err != nil {
		return err
	}
	temporary, err := os.MkdirTemp(filepath.Dir(current), ".activation-*")
	if err != nil {
		return err
	}
	defer os.RemoveAll(temporary)
	link := filepath.Join(temporary, "current")
	if err := os.Symlink(generation, link); err != nil {
		return fmt.Errorf("create generation link: %w", err)
	}
	if err := os.Rename(link, current); err != nil {
		return fmt.Errorf("activate generation (previous link retained): %w", err)
	}
	if err := syncActivationDirectory(filepath.Dir(current)); err != nil {
		return fmt.Errorf("generation activated, but directory sync failed: %w", err)
	}
	return nil
}

// GenerationDigest binds approval to every file, including downloads and SQL.
func GenerationDigest(directory string) (string, error) {
	hash := sha256.New()
	err := filepath.WalkDir(directory, func(path string, entry fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if entry.IsDir() {
			return nil
		}
		if !entry.Type().IsRegular() {
			return fmt.Errorf("generation contains non-regular file: %s", path)
		}
		relative, err := filepath.Rel(directory, path)
		if err != nil {
			return err
		}
		file, err := os.Open(path)
		if err != nil {
			return err
		}
		fileHash := sha256.New()
		_, copyErr := io.Copy(fileHash, file)
		closeErr := file.Close()
		if copyErr != nil {
			return copyErr
		}
		if closeErr != nil {
			return closeErr
		}
		fmt.Fprintf(hash, "%s\x00%x\n", filepath.ToSlash(relative), fileHash.Sum(nil))
		return nil
	})
	if err != nil {
		return "", err
	}
	return hex.EncodeToString(hash.Sum(nil)), nil
}
