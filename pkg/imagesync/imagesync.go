package imagesync

import (
	"archive/tar"
	"compress/gzip"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/yarlson/ftl/pkg/runner/remote"
)

// Config holds the configuration for the Docker image sync operation.
type Config struct {
	ImageName   string
	LocalStore  string
	RemoteStore string
	MaxParallel int
}

// ImageSync handles Docker image synchronization operations.
type ImageSync struct {
	cfg    Config
	runner *remote.Runner
}

// NewImageSync creates a new ImageSync instance with the provided configuration and SSH runner.
func NewImageSync(cfg Config, runner *remote.Runner) *ImageSync {
	if cfg.MaxParallel <= 0 {
		cfg.MaxParallel = 4
	}
	if cfg.LocalStore == "" {
		cfg.LocalStore = filepath.Join(os.Getenv("HOME"), "docker-images")
	}

	return &ImageSync{
		cfg:    cfg,
		runner: runner,
	}
}

// Sync performs the Docker image synchronization process.
func (s *ImageSync) Sync(ctx context.Context) error {
	needsSync, err := s.compareImages(ctx)
	if err != nil {
		return fmt.Errorf("failed to compare images: %w", err)
	}

	if !needsSync {
		return nil // Images are identical
	}

	if err := s.prepareDirectories(ctx); err != nil {
		return fmt.Errorf("failed to prepare directories: %w", err)
	}

	if err := s.exportAndExtractImage(ctx); err != nil {
		return fmt.Errorf("failed to export and extract image: %w", err)
	}

	if err := s.transferMetadata(ctx); err != nil {
		return fmt.Errorf("failed to transfer metadata: %w", err)
	}

	if err := s.syncBlobs(ctx); err != nil {
		return fmt.Errorf("failed to sync blobs: %w", err)
	}

	if err := s.loadRemoteImage(ctx); err != nil {
		return fmt.Errorf("failed to load remote image: %w", err)
	}

	return nil
}

// compareImages checks if the image needs to be synced by comparing local and remote versions.
func (s *ImageSync) compareImages(ctx context.Context) (bool, error) {
	localInspect, err := s.inspectLocalImage()
	if err != nil {
		return false, fmt.Errorf("failed to inspect local image: %w", err)
	}

	remoteInspect, err := s.inspectRemoteImage(ctx)
	if err != nil {
		return true, nil // Assume sync needed if remote inspection fails
	}

	// Compare normalized JSON data
	return !compareImageData(localInspect, remoteInspect), nil
}

// ImageData represents Docker image metadata.
type ImageData struct {
	Config struct {
		Hostname     string   `json:"Hostname"`
		Domainname   string   `json:"Domainname"`
		User         string   `json:"User"`
		AttachStdin  bool     `json:"AttachStdin"`
		AttachStdout bool     `json:"AttachStdout"`
		AttachStderr bool     `json:"AttachStderr"`
		ExposedPorts struct{} `json:"ExposedPorts"`
		Tty          bool     `json:"Tty"`
		OpenStdin    bool     `json:"OpenStdin"`
		StdinOnce    bool     `json:"StdinOnce"`
		Env          []string `json:"Env"`
		Cmd          []string `json:"Cmd"`
		Image        string   `json:"Image"`
		Volumes      struct{} `json:"Volumes"`
		WorkingDir   string   `json:"WorkingDir"`
		Entrypoint   []string `json:"Entrypoint"`
		OnBuild      []string `json:"OnBuild"`
		Labels       struct{} `json:"Labels"`
	} `json:"Config"`
	RootFS struct {
		Type    string   `json:"Type"`
		Layers  []string `json:"Layers"`
		DiffIDs []string `json:"DiffIDs"`
	} `json:"RootFS"`
	Architecture string `json:"Architecture"`
	Os           string `json:"Os"`
}

func (s *ImageSync) inspectLocalImage() (*ImageData, error) {
	cmd := exec.Command("docker", "inspect", s.cfg.ImageName)
	output, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("docker inspect failed: %w", err)
	}

	var data []ImageData
	if err := json.Unmarshal(output, &data); err != nil {
		return nil, fmt.Errorf("failed to parse inspect data: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no image data found")
	}

	return &data[0], nil
}

func (s *ImageSync) inspectRemoteImage(ctx context.Context) (*ImageData, error) {
	outputReader, err := s.runner.RunCommand(ctx, "docker", "inspect", s.cfg.ImageName)
	if err != nil {
		return nil, err
	}
	defer outputReader.Close()

	output, err := io.ReadAll(outputReader)
	if err != nil {
		return nil, fmt.Errorf("failed to read remote inspect data: %w", err)
	}

	var data []ImageData
	if err := json.Unmarshal([]byte(output), &data); err != nil {
		return nil, fmt.Errorf("failed to parse remote inspect data: %w", err)
	}

	if len(data) == 0 {
		return nil, fmt.Errorf("no remote image data found")
	}

	return &data[0], nil
}

func (s *ImageSync) prepareDirectories(ctx context.Context) error {
	if err := os.MkdirAll(s.cfg.LocalStore, 0755); err != nil {
		return fmt.Errorf("failed to create local store: %w", err)
	}

	if _, err := s.runner.RunCommand(ctx, fmt.Sprintf("mkdir -p %s", s.cfg.RemoteStore)); err != nil {
		return fmt.Errorf("failed to create remote store: %w", err)
	}

	return nil
}

func (s *ImageSync) exportAndExtractImage(ctx context.Context) error {
	imageDir := normalizeImageName(s.cfg.ImageName)
	localPath := filepath.Join(s.cfg.LocalStore, imageDir)

	if err := os.MkdirAll(localPath, 0755); err != nil {
		return fmt.Errorf("failed to create image directory: %w", err)
	}

	tarPath := filepath.Join(localPath, "image.tar")
	cmd := exec.Command("docker", "save", s.cfg.ImageName, "-o", tarPath)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to save image: %w", err)
	}

	if err := extractTar(ctx, tarPath, localPath); err != nil {
		return fmt.Errorf("failed to extract tar: %w", err)
	}

	return os.Remove(tarPath)
}

func extractTar(ctx context.Context, tarPath, destPath string) error {
	file, err := os.Open(tarPath)
	if err != nil {
		return fmt.Errorf("failed to open tar file: %w", err)
	}
	defer file.Close()

	var tr *tar.Reader

	// Check if the file is gzipped
	gzr, err := gzip.NewReader(file)
	if err == nil {
		defer gzr.Close()
		tr = tar.NewReader(gzr)
	} else {
		// If not gzipped, reset the file pointer and read as a regular tar
		if _, err := file.Seek(0, 0); err != nil {
			return fmt.Errorf("failed to reset file pointer: %w", err)
		}
		tr = tar.NewReader(file)
	}

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		header, err := tr.Next()
		if err == io.EOF {
			break // End of archive
		}
		if err != nil {
			return fmt.Errorf("tar reading error: %w", err)
		}

		target := filepath.Join(destPath, header.Name)

		switch header.Typeflag {
		case tar.TypeDir:
			if err := os.MkdirAll(target, 0755); err != nil {
				return fmt.Errorf("failed to create directory %s: %w", target, err)
			}
		case tar.TypeReg:
			if err := extractFile(tr, target); err != nil {
				return err
			}
		default:
			return fmt.Errorf("unsupported file type %b in %s", header.Typeflag, header.Name)
		}
	}

	return nil
}

func extractFile(tr *tar.Reader, target string) error {
	f, err := os.OpenFile(target, os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to create file %s: %w", target, err)
	}
	defer f.Close()

	if _, err := io.Copy(f, tr); err != nil {
		return fmt.Errorf("failed to write file %s: %w", target, err)
	}
	return nil
}

func (s *ImageSync) syncBlobs(ctx context.Context) error {
	localBlobs, err := s.listLocalBlobs()
	if err != nil {
		return err
	}

	remoteBlobs, err := s.listRemoteBlobs(ctx)
	if err != nil {
		return err
	}

	// Determine blobs to transfer
	var blobsToTransfer []string
	for _, blob := range localBlobs {
		if !contains(remoteBlobs, blob) {
			blobsToTransfer = append(blobsToTransfer, blob)
		}
	}

	// Transfer blobs in parallel batches
	return s.transferBlobs(ctx, blobsToTransfer)
}

func (s *ImageSync) transferBlobs(ctx context.Context, blobs []string) error {
	if len(blobs) == 0 {
		return nil
	}

	var wg sync.WaitGroup
	errChan := make(chan error, len(blobs))
	semaphore := make(chan struct{}, s.cfg.MaxParallel)

	for _, blob := range blobs {
		wg.Add(1)
		go func(blob string) {
			defer wg.Done()
			semaphore <- struct{}{}
			defer func() { <-semaphore }()

			if err := s.transferBlob(ctx, blob); err != nil {
				errChan <- fmt.Errorf("failed to transfer blob %s: %w", blob, err)
			}
		}(blob)
	}

	go func() {
		wg.Wait()
		close(errChan)
	}()

	for err := range errChan {
		if err != nil {
			return err
		}
	}

	return nil
}

func (s *ImageSync) loadRemoteImage(ctx context.Context) error {
	cmd := fmt.Sprintf("cd %s && tar -cf - . | docker load",
		filepath.Join(s.cfg.RemoteStore, normalizeImageName(s.cfg.ImageName)))

	outputReader, err := s.runner.RunCommand(ctx, cmd)
	if err != nil {
		return fmt.Errorf("failed to load remote image: %w", err)
	}
	defer outputReader.Close()

	_, err = io.ReadAll(outputReader)
	if err != nil {
		return fmt.Errorf("failed to read output of command '%s': %w", cmd, err)
	}

	return err
}

// Helper functions

func compareImageData(local, remote *ImageData) bool {
	localJSON, err := json.Marshal(local)
	if err != nil {
		return false
	}

	remoteJSON, err := json.Marshal(remote)
	if err != nil {
		return false
	}

	return string(localJSON) == string(remoteJSON)
}

func contains(slice []string, item string) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}

// listLocalBlobs returns a list of blob hashes from the local blob directory.
func (s *ImageSync) listLocalBlobs() ([]string, error) {
	imageDir := normalizeImageName(s.cfg.ImageName)
	blobPath := filepath.Join(s.cfg.LocalStore, imageDir, "blobs", "sha256")

	entries, err := os.ReadDir(blobPath)
	if err != nil {
		return nil, err
	}

	var blobs []string
	for _, entry := range entries {
		if !entry.IsDir() {
			blobs = append(blobs, entry.Name())
		}
	}

	return blobs, nil
}

// listRemoteBlobs returns a list of blob hashes from the remote blob directory.
func (s *ImageSync) listRemoteBlobs(ctx context.Context) ([]string, error) {
	imageDir := normalizeImageName(s.cfg.ImageName)
	blobPath := filepath.Join(s.cfg.RemoteStore, imageDir, "blobs", "sha256")

	output, err := s.runner.RunCommand(ctx, "ls", blobPath)
	if err != nil {
		return nil, nil // Return empty list if directory doesn't exist
	}

	data, err := io.ReadAll(output)
	if err != nil {
		return nil, err
	}

	blobs := strings.Fields(string(data))
	return blobs, nil
}

// transferBlob copies a single blob to the remote host.
func (s *ImageSync) transferBlob(ctx context.Context, blob string) error {
	imageDir := normalizeImageName(s.cfg.ImageName)
	localPath := filepath.Join(s.cfg.LocalStore, imageDir, "blobs", "sha256", blob)
	remotePath := filepath.Join(s.cfg.RemoteStore, imageDir, "blobs", "sha256", blob)

	_, err := s.runner.RunCommand(ctx, "mkdir", "-p", filepath.Dir(remotePath))
	if err != nil {
		return err
	}

	return s.runner.CopyFile(ctx, localPath, remotePath)
}

// transferMetadata copies the image metadata files to the remote host.
func (s *ImageSync) transferMetadata(ctx context.Context) error {
	imageDir := normalizeImageName(s.cfg.ImageName)
	localDir := filepath.Join(s.cfg.LocalStore, imageDir)
	remoteDir := filepath.Join(s.cfg.RemoteStore, imageDir)

	metadataFiles := []string{"index.json", "manifest.json", "oci-layout"}

	for _, file := range metadataFiles {
		localPath := filepath.Join(localDir, file)
		remotePath := filepath.Join(remoteDir, file)

		_, err := s.runner.RunCommand(ctx, "mkdir", "-p", filepath.Dir(remotePath))
		if err != nil {
			return err
		}

		if err := s.runner.CopyFile(ctx, localPath, remotePath); err != nil {
			return err
		}
	}

	return nil
}

func normalizeImageName(imageName string) string {
	imageName = strings.NewReplacer(":", "_", "/", "_").Replace(imageName)
	return imageName
}
