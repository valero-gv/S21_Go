package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
)

func handleError(err error) {
	if err != nil {
		log.Fatal(err)
	}
}

func main() {
	fileFlag := flag.Bool("f", false, "Print only files")
	dirFlag := flag.Bool("d", false, "Print only directories")
	symlinkFlag := flag.Bool("sl", false, "Print only symbolic links")
	extFlag := flag.String("ext", "", "Print only files with specified extension")
	flag.Parse()

	dirPath := flag.Arg(len(flag.Args()) - 1)
	if dirPath == "" {
		handleError(fmt.Errorf("directory path is required"))
	}
	stat, err := os.Stat(dirPath)
	if err != nil {
		handleError(err)
	}
	if !stat.IsDir() {
		handleError(fmt.Errorf("%s is not a directory", dirPath))
	}

	printAllTypes := !(*fileFlag || *dirFlag || *symlinkFlag)

	err = filepath.Walk(dirPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		switch {
		case (info.Mode().IsRegular() && *fileFlag) || (printAllTypes && info.Mode().IsRegular()):
			if *extFlag == "" || strings.HasSuffix(info.Name(), "."+*extFlag) {
				fmt.Println(path)
			}
		case (info.IsDir() && *dirFlag) || (printAllTypes && info.IsDir()):
			fmt.Println(path)
		case ((info.Mode()&os.ModeSymlink) != 0 && *symlinkFlag) || (printAllTypes && (info.Mode()&os.ModeSymlink) != 0):
			fmt.Println(path)
		}

		return nil
	})
	if err != nil {
		handleError(err)
	}
}
