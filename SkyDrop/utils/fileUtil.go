package utils

import (
	"errors"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"
)

func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsStatusFileExist(path string) bool {
	statusFilePath := GetStatusFilePath(path)
	return IsExist(statusFilePath)
}

func GetExeDir() string {
	exePath, err := os.Executable()
	if err != nil {
		panic(err)
	}
	return filepath.Dir(exePath)
}

func GetStatusFilePath(path string) string {
	// filename := filepath.Base(path)
	// fileDir := filepath.Dir(path)
	// return filepath.Join(fileDir, ".wait-del-"+filename)
	return filepath.Join(GetExeDir(), "tmp", strings.Replace(path, ":", "", 1))
}

func RemoveFileStatusFile(path string) {
	removeErr := os.Remove(path)
	if removeErr == nil {
		statusFilePath := GetStatusFilePath(path)
		os.Remove(statusFilePath)
	}
}

func CreateStatusFile(path string) {
	statusFilePath := GetStatusFilePath(path)
	fileDir := filepath.Dir(statusFilePath)
	os.MkdirAll(fileDir, 0755)
	file, err := os.Create(statusFilePath)
	if err != nil {
		log.Println(err)
		return
	}
	if runtime.GOOS == "windows" {
		// 设置文件为隐藏属性（Windows 特有的文件属性）
		//fileInfo, _ := file.Stat()
		//winAttr := uint32(fileInfo.Sys().(*syscall.Win32FileAttributeData).FileAttributes)
		//winAttr |= syscall.FILE_ATTRIBUTE_HIDDEN

		// 调用 syscall 设置新的文件属性
		//_ = syscall.SetFileAttributes(syscall.StringToUTF16Ptr(statusFilePath), winAttr)
		//SetFileHidden(statusFilePath)
	}
	file.Close()
}

func TimeString() string {
	return time.Now().Format("15:04")
}

func ReadSubDir(path string) ([]os.FileInfo, error) {

	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	dirs, err := f.Readdir(-1)
	if err != nil {
		return nil, err
	}

	return dirs, nil
}

func CurrentPath() string {
	dir, err := os.Getwd()
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("Current working directory:", dir)

	return dir
}

func GetCurrentPath() (string, error) {
	file, err := exec.LookPath(os.Args[0])
	if err != nil {
		return "", err
	}
	path, err := filepath.Abs(file)
	if err != nil {
		return "", err
	}
	i := strings.LastIndex(path, "/")
	if i < 0 {
		i = strings.LastIndex(path, "\\")
	}
	if i < 0 {
		return "", errors.New(`error: Can't find "/" or "\"`)
	}
	return string(path[0 : i+1]), nil
}

// CopyFile 复制文件从源路径到目标路径
func CopyFile(srcPath, dstPath string) error {
	// 打开源文件
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer srcFile.Close()

	// 创建目标文件
	dstFile, err := os.Create(dstPath)
	if err != nil {
		return err
	}
	defer dstFile.Close()

	// 复制文件内容
	_, err = io.Copy(dstFile, srcFile)
	if err != nil {
		return err
	}

	// 同步到磁盘
	err = dstFile.Sync()
	if err != nil {
		return err
	}

	// 复制文件权限
	srcInfo, err := os.Stat(srcPath)
	if err != nil {
		return err
	}
	err = os.Chmod(dstPath, srcInfo.Mode())
	if err != nil {
		return err
	}

	return nil
}

func FormatPathToWindowsSeparator(path string) string {
	return strings.Replace(path, "/", "\\", -1)
}

func FormatPathToUnixSeparator(path string) string {
	return strings.Replace(path, "\\", "/", -1)
}

func AddSeparator(path string) string {
	if !strings.HasSuffix(path, string(filepath.Separator)) {
		return path + string(filepath.Separator)
	}
	return path
}
