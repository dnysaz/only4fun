package handler

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strings"

	"only4.fun/backend/internal/builder"
	"only4.fun/backend/internal/database"

	"github.com/gin-gonic/gin"
)

const (
	// maxUploadBytes membatasi ukuran file ZIP yang diunggah (100 MB).
	maxUploadBytes = 100 << 20
	// maxExtractedBytes membatasi total ukuran setelah diekstrak agar ZIP
	// kecil tidak bisa mengisi disk (zip bomb).
	maxExtractedBytes = 500 << 20
	// maxZipEntries membatasi jumlah file dalam satu arsip.
	maxZipEntries = 20000
)

// zipRepoPrefix menandai project yang dibuat dari unggahan ZIP (bukan repo).
const zipRepoPrefix = "zip://"

// CreateFromZip membuat project baru dari arsip ZIP berisi file statis,
// publishes it, then returns the project.
func (h *ProjectHandler) CreateFromZip(c *gin.Context) {
	userID, ok := GetUserID(c)
	if !ok {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "not authenticated"})
		return
	}
	if h.sitesDir == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "sites folder is not configured"})
		return
	}

	c.Request.Body = http.MaxBytesReader(c.Writer, c.Request.Body, maxUploadBytes)
	if err := c.Request.ParseMultipartForm(16 << 20); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to read the upload (maximum 100 MB)"})
		return
	}

	subdomain := strings.ToLower(strings.TrimSpace(c.PostForm("subdomain")))
	if err := validateSubdomain(subdomain); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if existing, err := h.projectStore.GetBySubdomain(subdomain); err == nil && existing != nil {
		c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("project name %q is already taken", subdomain)})
		return
	}

	file, header, err := c.Request.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "ZIP file not found in the \"file\" field"})
		return
	}
	defer file.Close()
	if !strings.EqualFold(filepath.Ext(header.Filename), ".zip") {
		c.JSON(http.StatusBadRequest, gin.H{"error": "the file must have a .zip extension"})
		return
	}

	tmpZip, err := os.CreateTemp("", "upload-*.zip")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare a temporary file"})
		return
	}
	defer os.Remove(tmpZip.Name())
	defer tmpZip.Close()

	if _, err := io.Copy(tmpZip, file); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "failed to store the upload"})
		return
	}

	extractDir, err := os.MkdirTemp("", "site-upload-*")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare the extraction folder"})
		return
	}
	defer os.RemoveAll(extractDir)

	if err := extractZip(tmpZip.Name(), extractDir); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	root := stripSingleRoot(extractDir)
	if _, err := os.Stat(filepath.Join(root, "index.html")); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "the ZIP must contain index.html at its root (or inside a single folder)"})
		return
	}

	// Ekstraksi selesai dulu, baru project dibuat, supaya arsip yang rusak
	// tidak meninggalkan project kosong di dashboard.
	project, err := h.projectStore.Create(userID, zipRepoPrefix+header.Filename, subdomain, nil)
	if err != nil {
		if strings.Contains(strings.ToUpper(err.Error()), "UNIQUE") {
			c.JSON(http.StatusConflict, gin.H{"error": fmt.Sprintf("project name %q is already taken", subdomain)})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create project"})
		return
	}

	target := filepath.Join(h.sitesDir, subdomain)
	if err := builder.Publish(root, target); err != nil {
		log.Printf("failed to publish the ZIP upload for project %s: %v", project.ID, err)
		_ = h.projectStore.Delete(project.ID)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to publish the files"})
		return
	}
	release, _ := builder.CurrentReleaseName(target)

	if err := h.projectStore.UpdateStatus(project.ID, "deployed"); err != nil {
		log.Printf("failed to update project %s after upload: %v", project.ID, err)
	}
	if h.dns != nil && h.dns.Enabled() {
		if err := h.dns.EnsureRecord(c.Request.Context(), subdomain); err != nil {
			log.Printf("failed to ensure DNS for %s after upload: %v", subdomain, err)
		}
	}

	if h.deploymentStore != nil {
		record, err := h.deploymentStore.Create(project.ID, "success", fmt.Sprintf("ZIP upload: %s (%d files)", header.Filename, countFiles(root)))
		if err == nil && release != "" {
			_ = h.deploymentStore.SetRelease(record.ID, release)
		}
	}

	project.Status = "deployed"
	project.EnvKeys = database.ParseEnvKeys(project.EnvVars)
	c.JSON(http.StatusCreated, gin.H{"project": project})
}

// extractZip mengekstrak arsip dengan proteksi path traversal, symlink, dan
// batas ukuran.
func extractZip(zipPath, dst string) error {
	reader, err := zip.OpenReader(zipPath)
	if err != nil {
		return fmt.Errorf("invalid ZIP file")
	}
	defer reader.Close()

	if len(reader.File) > maxZipEntries {
		return fmt.Errorf("archive contains too many files (maximum %d)", maxZipEntries)
	}

	root := filepath.Clean(dst)
	var total int64
	for _, entry := range reader.File {
		name := strings.ReplaceAll(entry.Name, "\\", "/")
		// path.Clean dengan prefix "/" menormalkan "../" sebelum digabung.
		rel := strings.TrimPrefix(path.Clean("/"+name), "/")
		if rel == "" || rel == "." {
			continue
		}
		target := filepath.Join(root, filepath.FromSlash(rel))
		if target != root && !strings.HasPrefix(target, root+string(os.PathSeparator)) {
			return fmt.Errorf("archive contains an invalid path: %s", entry.Name)
		}

		if entry.FileInfo().IsDir() {
			if err := os.MkdirAll(target, 0o755); err != nil {
				return fmt.Errorf("failed to create a folder from the archive")
			}
			continue
		}
		// Lewati symlink dan entri khusus; hanya file biasa yang diekstrak.
		if !entry.Mode().IsRegular() {
			continue
		}

		if total+int64(entry.UncompressedSize64) > maxExtractedBytes {
			return fmt.Errorf("archive contents are too large after extraction")
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return fmt.Errorf("failed to create a folder from the archive")
		}

		src, err := entry.Open()
		if err != nil {
			return fmt.Errorf("failed to read archive contents")
		}
		out, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o644)
		if err != nil {
			src.Close()
			return fmt.Errorf("failed to write a file from the archive")
		}
		// Batasi juga saat menyalin: UncompressedSize64 di header bisa dipalsukan,
		// jadi jumlah byte nyata yang dihitung.
		written, copyErr := io.Copy(out, io.LimitReader(src, maxExtractedBytes-total+1))
		src.Close()
		out.Close()
		total += written
		if copyErr != nil {
			return fmt.Errorf("failed to copy a file from the archive")
		}
		if total > maxExtractedBytes {
			return fmt.Errorf("archive contents are too large after extraction")
		}
	}

	return nil
}

// stripSingleRoot returns the contents of a folder if the archive wraps everything
// inside a single folder (e.g., project/index.html); otherwise the folder itself.
func stripSingleRoot(dir string) string {
	entries, err := os.ReadDir(dir)
	if err != nil || len(entries) != 1 || !entries[0].IsDir() {
		return dir
	}
	inner := filepath.Join(dir, entries[0].Name())
	if _, err := os.Stat(filepath.Join(inner, "index.html")); err == nil {
		return inner
	}
	return dir
}

// countFiles menghitung file di folder (dipakai untuk log deployment).
func countFiles(dir string) int {
	count := 0
	_ = filepath.Walk(dir, func(_ string, info os.FileInfo, err error) error {
		if err == nil && info != nil && info.Mode().IsRegular() {
			count++
		}
		return nil
	})
	return count
}
