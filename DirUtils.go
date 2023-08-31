package main

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

func getCurrentDir() (string, []string) {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Println(err)
	}
	var contents []string
	for _, file := range files {
		if file.IsDir() {
			contents = append(contents, "📁 "+file.Name())
		} else {
			contents = append(contents, "📄 "+file.Name())
		}
	}
	return dir, contents
}

func showDir(path string) string {
	files, err := os.ReadDir(path)
	if err != nil {
		fmt.Println(err)
	}

	var result string
	for _, file := range files {
		if file.IsDir() {
			result += fmt.Sprintf("📁 %s\n", file.Name())
			result += showDir(filepath.Join(path, file.Name()))
		} else {
			result += fmt.Sprintf("📄 %s\n", file.Name())
		}
	}
	return result
}

func CopyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()

	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()

	_, err = io.Copy(out, in)
	if err != nil {
		return err
	}

	return out.Close()
}

func CopyDir(src, dst string) error {
	info, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !info.IsDir() {
		return CopyFile(src, dst)
	}

	// Create the destination directory with the same name as the source directory
	dst = filepath.Join(dst, filepath.Base(src))
	err = os.MkdirAll(dst, info.Mode())
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(src)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		srcPath := filepath.Join(src, entry.Name())
		dstPath := filepath.Join(dst, entry.Name())

		if entry.IsDir() {
			err = CopyDir(srcPath, dstPath)
			if err != nil {
				return err
			}
		} else {
			err = CopyFile(srcPath, dstPath)
			if err != nil {
				return err
			}
		}
	}

	return nil
}

func removePath(path string) error {
	err := os.RemoveAll(path)
	if err != nil {
		return err
	}
	return nil
}

func createNewDir(targetPath string, newDirName string) error {
	newDirPath := filepath.Join(targetPath, newDirName)
	err := os.MkdirAll(newDirPath, 0755)
	if err != nil {
		return err
	}

	return nil
}
