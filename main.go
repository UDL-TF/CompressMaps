package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/dsnet/compress/bzip2"
)

const (
	splitThresholdBytes = 26214400         // 25 MiB
	splitChunkSize      = 20 * 1024 * 1024 // 20 MiB

	// ANSI color codes
	colorRed    = "\033[0;31m"
	colorGreen  = "\033[0;32m"
	colorYellow = "\033[1;33m"
	colorReset  = "\033[0m"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "%sError: %v%s\n", colorRed, err, colorReset)
		os.Exit(1)
	}
}

func run() error {
	// Check if file argument provided
	if len(os.Args) < 2 {
		fmt.Fprintf(os.Stderr, "%sError: No map file provided%s\n", colorRed, colorReset)
		fmt.Println("Usage: compressmaps <map_file.bsp>")
		fmt.Println("Or drag and drop a .bsp file onto this executable")
		return fmt.Errorf("no file provided")
	}

	mapFile := os.Args[1]

	// Validate file exists
	if _, err := os.Stat(mapFile); os.IsNotExist(err) {
		return fmt.Errorf("%sfile not found: %s%s", colorRed, mapFile, colorReset)
	}

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(mapFile), ".bsp") {
		return fmt.Errorf("%sfile must have .bsp extension%s", colorRed, colorReset)
	}

	// Get absolute path
	absPath, err := filepath.Abs(mapFile)
	if err != nil {
		return fmt.Errorf("failed to get absolute path: %w", err)
	}

	targetPath := absPath + ".bz2"
	partsDir := targetPath + ".parts"

	// Check if output already exists
	if fileExists(targetPath) || dirExists(partsDir) {
		fmt.Printf("%sWarning: Output already exists for %s%s\n", colorYellow, filepath.Base(mapFile), colorReset)
		fmt.Print("Overwrite? (y/N): ")
		reader := bufio.NewReader(os.Stdin)
		response, _ := reader.ReadString('\n')
		response = strings.TrimSpace(strings.ToLower(response))

		if response != "y" && response != "yes" {
			fmt.Println("Cancelled")
			return nil
		}

		// Remove existing files
		os.Remove(targetPath)
		os.RemoveAll(partsDir)
	}

	fmt.Printf("%sCompressing %s to bzip2...%s\n", colorGreen, filepath.Base(mapFile), colorReset)

	// Compress the file
	if err := compressFile(absPath, targetPath); err != nil {
		return fmt.Errorf("compression failed: %w", err)
	}
	fmt.Printf("%s✓ Created %s%s\n", colorGreen, targetPath, colorReset)

	// Check file size
	fileInfo, err := os.Stat(targetPath)
	if err != nil {
		return fmt.Errorf("failed to stat compressed file: %w", err)
	}
	fileSize := fileInfo.Size()
	fmt.Printf("Compressed size: %d MiB\n", fileSize/1024/1024)

	// Split if needed
	if fileSize > splitThresholdBytes {
		fmt.Printf("%sFile exceeds %d bytes threshold%s\n", colorYellow, splitThresholdBytes, colorReset)
		fmt.Printf("%sSplitting into 20 MiB chunks...%s\n", colorGreen, colorReset)

		if err := os.MkdirAll(partsDir, 0755); err != nil {
			return fmt.Errorf("failed to create parts directory: %w", err)
		}

		baseName := filepath.Base(targetPath)
		if err := splitFile(targetPath, partsDir, baseName, splitChunkSize); err != nil {
			return fmt.Errorf("splitting failed: %w", err)
		}

		// Remove the original compressed file
		if err := os.Remove(targetPath); err != nil {
			return fmt.Errorf("failed to remove original compressed file: %w", err)
		}

		// Count parts
		entries, err := os.ReadDir(partsDir)
		if err != nil {
			return fmt.Errorf("failed to read parts directory: %w", err)
		}
		partCount := len(entries)
		fmt.Printf("%s✓ Split into %d part(s) in %s%s\n", colorGreen, partCount, partsDir, colorReset)

		// List the parts
		fmt.Println("Parts created:")
		for _, entry := range entries {
			info, err := entry.Info()
			if err != nil {
				continue
			}
			fmt.Printf("  %s (%d bytes)\n", entry.Name(), info.Size())
		}
	} else {
		fmt.Printf("%s✓ File size is within threshold, no splitting needed%s\n", colorGreen, colorReset)
	}

	fmt.Printf("%sDone!%s\n", colorGreen, colorReset)

	return nil
}

func compressFile(inputPath, outputPath string) error {
	// Open input file
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	// Create output file
	outputFile, err := os.Create(outputPath)
	if err != nil {
		return fmt.Errorf("failed to create output file: %w", err)
	}
	defer outputFile.Close()

	// Create bzip2 writer
	bz2Writer, err := bzip2.NewWriter(outputFile, &bzip2.WriterConfig{Level: 9})
	if err != nil {
		return fmt.Errorf("failed to create bzip2 writer: %w", err)
	}
	defer bz2Writer.Close()

	// Copy and compress
	if _, err := io.Copy(bz2Writer, inputFile); err != nil {
		return fmt.Errorf("failed to compress: %w", err)
	}

	return nil
}

func splitFile(inputPath, outputDir, baseName string, chunkSize int64) error {
	inputFile, err := os.Open(inputPath)
	if err != nil {
		return fmt.Errorf("failed to open input file: %w", err)
	}
	defer inputFile.Close()

	partNum := 0
	buffer := make([]byte, chunkSize)

	for {
		// Read chunk
		n, err := io.ReadFull(inputFile, buffer)
		if err != nil && err != io.EOF && err != io.ErrUnexpectedEOF {
			return fmt.Errorf("failed to read chunk: %w", err)
		}

		if n == 0 {
			break
		}

		// Create part file
		partFileName := fmt.Sprintf("%s.part.%03d", baseName, partNum)
		partPath := filepath.Join(outputDir, partFileName)
		partFile, err := os.Create(partPath)
		if err != nil {
			return fmt.Errorf("failed to create part file: %w", err)
		}

		// Write chunk
		if _, err := partFile.Write(buffer[:n]); err != nil {
			partFile.Close()
			return fmt.Errorf("failed to write part file: %w", err)
		}
		partFile.Close()

		partNum++

		if err == io.EOF || err == io.ErrUnexpectedEOF {
			break
		}
	}

	return nil
}

func fileExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return !info.IsDir()
}

func dirExists(path string) bool {
	info, err := os.Stat(path)
	if os.IsNotExist(err) {
		return false
	}
	return info.IsDir()
}
