package main

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

func addFileToTar(filename string, file *os.File, modTime time.Time, tw *tar.Writer) error {
	fileInfo, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := tar.FileInfoHeader(fileInfo, filename)
	if err != nil {
		return err
	}

	header.ModTime = modTime

	err = tw.WriteHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(tw, file)
	if err != nil {
		return err
	}

	return nil
}

func rotate(logfile, archiveDir string) error {
	fileInfo, err := os.Stat(logfile)
	if os.IsNotExist(err) {
		return fmt.Errorf("file %s does not exist", logfile)
	}
	modTime := fileInfo.ModTime()
	archiveName := filepath.Base(logfile) + "_" + modTime.Format("20060102150405") + ".tar.gz"
	file, err := os.Open(logfile)
	if err != nil {
		return fmt.Errorf("failed to open file %s: %v", logfile, err)
	}
	defer file.Close()

	archivePath := archiveName
	if archiveDir != "" {
		archivePath = filepath.Join(archiveDir, archiveName)
	}
	archiveFile, err := os.Create(archivePath)
	if err != nil {
		return fmt.Errorf("failed to create archive file: %v", err)
	}
	defer archiveFile.Close()
	gw := gzip.NewWriter(archiveFile)
	defer gw.Close()
	tw := tar.NewWriter(gw)
	defer tw.Close()
	err = addFileToTar(logfile, file, modTime, tw)
	if err != nil {
		return fmt.Errorf("failed to add file to archive: %v", err)
	}

	fmt.Printf("Rotated %s to %s\n", logfile, archivePath)
	return nil
}

func main() {
	args := os.Args[1:]
	if len(args) == 0 {
		fmt.Println("Usage: myRotate [-a archive_dir] logfile1 [logfile2 ...]")
		os.Exit(1)
	}
	var archiveDir string
	if args[0] == "-a" {
		if len(args) < 0 {
			fmt.Println("Usage: myRotate [-a archive_dir] logfile1 [logfile2 ...]")
			os.Exit(1)
		}
		archiveDir = args[1]
		args = args[2:]
	}
	var wg sync.WaitGroup
	for _, logfile := range args {
		wg.Add(1)
		go func(logfile string) {
			defer wg.Done()
			err := rotate(logfile, archiveDir)
			if err != nil {
				fmt.Printf("Error rotating %s: %v\n", logfile, err)
			}
		}(logfile)
	}
	wg.Wait()
}
