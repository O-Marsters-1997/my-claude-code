package main

import (
	"os"
	"path/filepath"
	"strings"
)

func mainCheckout(dir string) string {
	for d := dir; ; {
		gitPath := filepath.Join(d, ".git")
		if info, err := os.Stat(gitPath); err == nil {
			if info.IsDir() {
				return d
			}
			return linkedMain(gitPath, dir)
		}
		parent := filepath.Dir(d)
		if parent == d {
			return dir
		}
		d = parent
	}
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
