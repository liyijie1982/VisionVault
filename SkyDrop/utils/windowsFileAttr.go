//go:build windows
// +build windows

package utils

import (
	"os"
	"syscall"
)

// setFileHiddenWindows 是 Windows 平台下的实现
// 使用 build constraints 限制该函数仅在 Windows 下编译
func SetFileHidden(filePath string) error {
	// 引入 syscall 包（仅在 Windows 下有效）
	_, err := os.Stat(filePath)
	if err != nil {
		return err
	}

	// 调用 Windows API 设置文件属性
	err = syscall.SetFileAttributes(syscall.StringToUTF16Ptr(filePath), syscall.FILE_ATTRIBUTE_HIDDEN)
	if err != nil {
		return err
	}
	return nil
}
