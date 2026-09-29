package logstore

import (
	"os"
	"path/filepath"
	"strings"
)

func CheckoutRoot(dir string) string {
	for d := dir; ; {
		if _, err := os.Stat(filepath.Join(d, ".git")); err == nil {
			return d
		}
		parent := filepath.Dir(d)
		if parent == d {
			return dir
		}
		d = parent
	}
}

func mainCheckout(dir string) string {
	root := CheckoutRoot(dir)
	gitPath := filepath.Join(root, ".git")
	if info, err := os.Stat(gitPath); err != nil || info.IsDir() {
		return root
	}
	return linkedMain(gitPath, root)
}

func linkedMain(gitFile, fallback string) string {
	b, err := os.ReadFile(gitFile)
	if err != nil {
		return fallback
	}
	gitdir := strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(string(b)), "gitdir:"))
	if !filepath.IsAbs(gitdir) {
		gitdir = filepath.Join(filepath.Dir(gitFile), gitdir)
	}
	main, _, found := strings.Cut(filepath.ToSlash(gitdir), "/.git/")
	if !found {
		return fallback
	}
	return filepath.FromSlash(main)
}
