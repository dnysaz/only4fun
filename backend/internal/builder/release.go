package builder

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Publish mempublikasikan folder hasil build/ekstraksi secara atomik ke
// <sitesDir>/<subdomain> (symlink ke folder rilis baru).
func Publish(src, dst string) error {
	return publish(src, dst)
}

// CurrentReleaseName mengembalikan nama folder rilis yang sedang aktif untuk
// sebuah situs (target adalah <sitesDir>/<subdomain>).
func CurrentReleaseName(target string) (string, error) {
	resolved, err := os.Readlink(target)
	if err != nil {
		return "", fmt.Errorf("site is not using a release folder yet: %w", err)
	}
	return filepath.Base(resolved), nil
}

// CleanUnusedReleases menghapus folder rilis yang tidak lagi dipakai, kecuali
// rilis yang saat ini menjadi target symlink situs. Dipanggil setelah history
// deployment dihapus: rilis lama tidak bisa di-rollback lagi, tapi situs yang
// sedang live harus tetap utuh.
func CleanUnusedReleases(target string) error {
	active, err := CurrentReleaseName(target)
	if err != nil {
		active = "" // Situs bukan symlink: tidak ada rilis yang dipakai.
	}
	parent := filepath.Dir(target)
	releasesDir := filepath.Join(parent, ".releases", filepath.Base(target))
	entries, err := os.ReadDir(releasesDir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		if name == active || name != filepath.Base(name) || strings.Contains(name, "..") {
			continue
		}
		if err := os.RemoveAll(filepath.Join(releasesDir, name)); err != nil {
			return err
		}
	}
	return nil
}

// ActivateRelease mengarahkan symlink situs ke folder rilis lama sehingga
// deployment sebelumnya bisa dipromosikan kembali (rollback).
//
// release harus berupa nama folder di .releases/<subdomain>; nilai dengan
// pemisah path ditolak agar tidak bisa keluar dari folder rilis.
func ActivateRelease(target, release string) error {
	if release == "" || release != filepath.Base(release) || strings.Contains(release, "..") {
		return fmt.Errorf("invalid release name")
	}

	parent := filepath.Dir(target)
	base := filepath.Base(target)
	releasePath := filepath.Join(parent, ".releases", base, release)
	info, err := os.Stat(releasePath)
	if err != nil || !info.IsDir() {
		return fmt.Errorf("release %q was not found", release)
	}

	linkTmp := filepath.Join(parent, fmt.Sprintf(".link-%s-%d", base, time.Now().UnixNano()))
	_ = os.Remove(linkTmp)
	releaseAbs, err := filepath.Abs(releasePath)
	if err != nil {
		return fmt.Errorf("failed to resolve the release folder: %w", err)
	}
	if err := os.Symlink(releaseAbs, linkTmp); err != nil {
		return fmt.Errorf("failed to prepare the release symlink: %w", err)
	}
	if err := os.Rename(linkTmp, target); err != nil {
		_ = os.Remove(linkTmp)
		return fmt.Errorf("failed to activate the release: %w", err)
	}
	return nil
}
