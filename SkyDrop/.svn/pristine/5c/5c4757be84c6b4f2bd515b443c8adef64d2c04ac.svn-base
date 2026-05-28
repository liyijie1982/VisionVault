package utils

import (
	"errors"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
	"strconv"
)

func WalkDir(root, path string, conn net.Conn) error {
	infos, err := ReadSubDir(path)
	if err != nil {
		return err
	}
	for _, info := range infos {
		_filePath := filepath.Join(path, info.Name())
		if info.IsDir() {
			err = WalkDir(root, _filePath, conn)
		} else {
			err = SendFile(info, root, _filePath, conn)
		}
	}
	return err
}

func SendFileData(path string, conn net.Conn) error {
	// 以只读方式打开文件
	f, err := os.Open(path)
	if err != nil {
		fmt.Println("os.Open err:", err)
		return err
	}
	defer f.Close() // 发送结束关闭文件。

	// 循环读取文件，原封不动的写给服务器
	buf := make([]byte, 4096)
	for {
		n, err := f.Read(buf) // 读取文件内容到切片缓冲中
		if err != nil {
			if err == io.EOF {
				fmt.Printf("%s, 文件发送完毕\n", path)
			} else {
				fmt.Println("f.Read err:", err)
			}
			break
		}
		_, err = conn.Write(buf[:n]) // 原封不动写给服务器
	}
	return err
}

func SyncFile(path, tcpServerAddr string, isDel bool) error {
	// 主动连接服务器
	conn, err := net.Dial("tcp", tcpServerAddr)
	if err != nil {
		fmt.Println("net.Dial err:", err)
		return err
	}
	defer conn.Close()

	fileInfo, err := os.Stat(path)
	if err != nil {
		fmt.Println("os.Stat err:", err)
		return err
	}

	if fileInfo.IsDir() {
		err = WalkDir(path, path, conn)
	} else {
		err = SendFile(fileInfo, path, path, conn)
	}

	if err == nil && isDel {
		err = os.RemoveAll(path)
	}
	return err
}

func SendFile(fileInfo os.FileInfo, root, path string, conn net.Conn) error {
	if SendFileHeader(fileInfo, root, path, conn) {
		return SendFileData(path, conn)
	} else {
		return errors.New("SendFileHeader err")
	}
}

func SendFileHeader(fileInfo os.FileInfo, root, path string, conn net.Conn) bool {

	relPath, _ := filepath.Rel(root, path)
	rootName := filepath.Base(root)
	relPath = filepath.Join(rootName, relPath)
	// 给接收端，先发送文件名
	SendData([]byte(relPath), conn)
	if RecvOK(conn) {
		strSize := strconv.FormatInt(fileInfo.Size(), 10)
		SendData([]byte(strSize), conn)
		return RecvOK(conn)
	}
	return false
}

func SendData(data []byte, conn net.Conn) {
	_, err := conn.Write(data)
	if err != nil {
		fmt.Println("conn.Write err:", err)
		return
	}
}

func RecvOK(conn net.Conn) bool {
	// 读取接收端回发确认数据 —— ok
	buf := make([]byte, 1024)
	n, err := conn.Read(buf)
	if err != nil {
		fmt.Println("conn.Read err:", err)
		return false
	}
	return "ok" == string(buf[:n])
}
