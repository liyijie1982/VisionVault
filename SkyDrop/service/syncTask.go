package service

import (
	"SyncAgent/utils"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"time"
)

type SyncFileInfo struct {
	Path       string
	Size       int64
	ChangeTime int64
}

type SyncFileRlt struct {
	Path       string
	DstPath    string
	FileSize   int64
	Err        error
	ChangeTime string
}

func FileSyncTask(storage StorageConfig, srcRootPath string, maxWorkers int) error {
	defer func() { //捕获异常
		if r := recover(); r != nil {
			fmt.Println("发生异常:", r)
		}
	}()

	taskStartTime := utils.GetMillisecond()

	fileQueues := make(chan SyncFileInfo, maxWorkers*2) //任务队列
	statQueues := make(chan SyncFileRlt, maxWorkers*2)  //任务队列

	//目录扫描
	log.Println("ScanFileAndSendTask Start ......")
	go ScanFileAndSendTask(srcRootPath, srcRootPath, fileQueues)

	var waitUploadWorker sync.WaitGroup //上传任务完成状态信号
	for i := 0; i < maxWorkers; i++ {
		s3Session, err := InitS3Conn(storage)
		if err != nil {
			return err
		}

		waitUploadWorker.Add(1)
		if storage.Type == "s3" {
			go S3UploadWorker(fileQueues, statQueues, s3Session, storage.Bucket, srcRootPath, &waitUploadWorker)
		} else { //local -本地存储
			go LocalUploadWorker(fileQueues, statQueues, storage.LocalPath, srcRootPath, &waitUploadWorker)
		}
	}

	go func() {
		waitUploadWorker.Wait()
		close(statQueues)
		log.Println("uploadWorker finished", srcRootPath)
	}()

	return doSyncResult(statQueues, srcRootPath, taskStartTime, storage)
}

func doSyncResult(syncFileRlt chan SyncFileRlt, srcRootPath, taskStartTime string, storage StorageConfig) error {

	logFileName := NewConstants().LogFilePath(srcRootPath)
	fileDir := filepath.Dir(logFileName)
	os.MkdirAll(fileDir, 0755)
	logFile, openErr := os.OpenFile(logFileName, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0666)
	if openErr != nil {
		return openErr
	}
	defer func() {
		_ = logFile.Close()
		_ = os.Remove(logFileName)
	}()

	var fileSize, fileCount, errCount int64
	for stat := range syncFileRlt {
		fileSize = fileSize + stat.FileSize
		fileCount++

		errString := ""
		if stat.Err != nil {
			errString = stat.Err.Error()
			errCount++
		}

		data := map[string]string{
			"path":       stat.Path,
			"dstPath":    stat.DstPath,
			"size":       formatFileSize(stat.FileSize),
			"changeTime": stat.ChangeTime,
			"uploadTime": time.Now().Format("2006-01-02 15:04:05"),
			"error":      errString,
		}
		jsonData, err := json.Marshal(data)
		if err != nil {
			return err
		}
		if _, err = io.WriteString(logFile, string(jsonData)+"\r\n"); err != nil {
			return err
		}
	}

	if fileCount == 0 {
		log.Println("skip sync commit because no files were uploaded", srcRootPath)
		return nil
	}

	logPath := ""
	if (errCount > 0 || fileCount > 0) && storage.Type == "s3" {
		logPath, _ = uploadLogFile(logFileName, storage)
	} else {
		logPath = logFileName
	}

	return commitTask(srcRootPath, taskStartTime, logPath, fileSize, fileCount, errCount)
}

func commitTask(srcRootPath, taskStartTime, tarListPath string, fileSize, fileCount, errCount int64) error {
	params := url.Values{}
	params.Add("ip", NewConstants().IP)
	params.Add("path", srcRootPath)
	params.Add("fileSize", strconv.FormatInt(fileSize, 10))
	params.Add("fileCount", strconv.FormatInt(fileCount, 10))
	params.Add("errCount", strconv.FormatInt(errCount, 10))
	params.Add("taskStartTime", taskStartTime)
	params.Add("tarListPath", tarListPath)
	_, err := utils.HttpGet(NewConstants().Console+"/sky/agent/commit", params.Encode())
	if err != nil {
		log.Println(err.Error())
	}
	return err
}

func uploadLogFile(logFileName string, storage StorageConfig) (string, error) {
	s3Session, err := InitS3Conn(storage)
	if err != nil {
		return "", err
	}
	return S3UploadLogFile(logFileName, storage.Bucket, s3Session)
}
