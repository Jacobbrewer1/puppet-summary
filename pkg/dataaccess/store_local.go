package dataaccess

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/prometheus/client_golang/prometheus"
)

type localImpl struct {
	// reportsDir is the directory to store reports in.
	reportsDir string
}

func newLocal() (*localImpl, error) {
	pwd, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("error getting current working directory: %w", err)
	}

	// Get the "reports" directory.
	reportsDir := filepath.Join(pwd, "reports")

	// Get the full path to the "reports" directory.
	repsDir, err := filepath.Abs(reportsDir)
	if err != nil {
		return nil, fmt.Errorf("error getting absolute path to dumps directory: %w", err)
	}

	return &localImpl{
		reportsDir: repsDir,
	}, nil
}

func (l *localImpl) SaveFile(_ context.Context, filePath string, file []byte) error {
	// Start the prometheus timer.
	t := prometheus.NewTimer(StorageLatency.With(prometheus.Labels{"query": "save_file"}))
	defer t.ObserveDuration()

	// Add the reports directory to the file path.
	filePath = filepath.Join(l.reportsDir, filePath)

	// Create the directory if it does not exist.
	err := os.MkdirAll(filepath.Dir(filePath), os.ModePerm)
	if err != nil {
		return fmt.Errorf("error creating directory: %w", err)
	}

	// Create a new file.
	w, err := os.Create(filePath)
	if err != nil {
		return fmt.Errorf("error creating file: %w", err)
	}

	// Write the file to the bucket.
	_, err = w.Write(file)
	if err != nil {
		return fmt.Errorf("error writing file to directory: %w", err)
	}

	// Close the file.
	err = w.Close()
	if err != nil {
		return fmt.Errorf("error closing file: %w", err)
	}

	return nil
}

func (l *localImpl) DownloadFile(_ context.Context, filePath string) ([]byte, error) {
	// Start the prometheus timer.
	t := prometheus.NewTimer(StorageLatency.With(prometheus.Labels{"query": "download_file"}))
	defer t.ObserveDuration()

	// Add the reports directory to the file path.
	filePath = filepath.Join(l.reportsDir, filePath)

	// Open the file.
	r, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("error opening file: %w", err)
	}

	// Read the file.
	file, err := io.ReadAll(r)
	if err != nil {
		return nil, fmt.Errorf("error reading file: %w", err)
	}

	// Close the file.
	err = r.Close()
	if err != nil {
		return nil, fmt.Errorf("error closing file: %w", err)
	}

	return file, nil
}

func (l *localImpl) DeleteFile(_ context.Context, filePath string) error {
	// Start the prometheus timer.
	t := prometheus.NewTimer(StorageLatency.With(prometheus.Labels{"query": "delete_file"}))
	defer t.ObserveDuration()

	// Add the reports directory to the file path.
	filePath = filepath.Join(l.reportsDir, filePath)

	// Delete the file.
	err := os.Remove(filePath)
	if err != nil {
		return fmt.Errorf("error deleting file: %w", err)
	}

	return nil
}

func (l *localImpl) Purge(_ context.Context, from time.Time) (int, error) {
	// Start the prometheus timer.
	t := prometheus.NewTimer(StorageLatency.With(prometheus.Labels{"query": "purge"}))
	defer t.ObserveDuration()

	// Read the files in the directory.
	files, err := os.ReadDir(l.reportsDir)
	if err != nil {
		return 0, fmt.Errorf("error reading directory: %w", err)
	}

	// Purge the files.
	count := 0
	for _, file := range files {
		// Here we are at the environment level
		// We need to go into the environment and then into the node

		// Get the environment directory.
		envDir := filepath.Join(l.reportsDir, file.Name())
		envFiles, err := os.ReadDir(envDir)
		if err != nil {
			return 0, fmt.Errorf("error reading directory: %w", err)
		}

		// Purge the files.
		for _, envFile := range envFiles {
			// Now we are at the Node level
			// We need to go into the node and then into the report

			// Get the node directory.
			nodeDir := filepath.Join(envDir, envFile.Name())
			nodeFiles, err := os.ReadDir(nodeDir)
			if err != nil {
				return 0, fmt.Errorf("error reading directory: %w", err)
			}

			// Purge the files.
			for _, nodeFile := range nodeFiles {
				// Now we are at the report level
				// We can now purge the report

				// Get the report file.
				reportFile := filepath.Join(nodeDir, nodeFile.Name())

				// Remove the path (report/environment/fqdn) from the file name.
				reportFile = reportFile[strings.LastIndex(reportFile, "/")+1:]

				// Remove the file extension.
				reportFile = reportFile[:len(reportFile)-len(".yaml")]

				// Get the timestamp of the report.
				timestamp, err := time.Parse(time.RFC3339, reportFile)
				if err != nil {
					return 0, fmt.Errorf("error parsing file name: %w", err)
				}

				// Check if the file is older than the purge date.
				if timestamp.After(from) {
					continue
				}

				// Delete the file.
				err = os.Remove(filepath.Join(nodeDir, nodeFile.Name()))
				if err != nil {
					return 0, fmt.Errorf("error deleting file: %w", err)
				}

				count++
			}

			err = checkEmptyDir(nodeDir)
			if err != nil {
				return 0, fmt.Errorf("error checking empty directory: %w", err)
			}
		}

		err = checkEmptyDir(envDir)
		if err != nil {
			return 0, fmt.Errorf("error checking empty directory: %w", err)
		}
	}

	err = checkEmptyDir(l.reportsDir)
	if err != nil {
		return 0, fmt.Errorf("error checking empty directory: %w", err)
	}

	return count, nil
}

// checkEmptyDir checks if a directory is empty and deletes it if it is.
func checkEmptyDir(path string) error {
	dir, err := os.ReadDir(path)
	if err != nil {
		return fmt.Errorf("error reading directory: %w", err)
	}

	if len(dir) == 0 {
		err = os.Remove(path)
		if err != nil {
			return fmt.Errorf("error deleting directory: %w", err)
		}
	}

	return nil
}
