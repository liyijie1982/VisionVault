package service

import (
	"SyncAgent/utils"
	"fmt"
	"log"
	"math"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
)

func S3Conn(endPoint, accessKey, secretKey string) (*session.Session, error) {
	log.Printf("endPoint: %s", endPoint)
	log.Printf("ak: %s, sk: %s", accessKey, "******")

	return session.NewSession(&aws.Config{
		Credentials:      credentials.NewStaticCredentials(accessKey, secretKey, ""),
		Endpoint:         aws.String(endPoint),
		Region:           aws.String("us-east-1"),
		DisableSSL:       aws.Bool(true),
		S3ForcePathStyle: aws.Bool(true), //virtual-host style方式，不要修改
	})
}

func InitS3Conn(storage StorageConfig) (*session.Session, error) {
	s3Session, err := S3Conn(storage.Endpoint, storage.AK, storage.SK)
	if err != nil {
		log.Println(err.Error())
		return nil, err
	}
	return s3Session, nil
}

func S3UploadWorker(fileQueues chan SyncFileInfo,
	syncFileRlt chan SyncFileRlt,
	s3Session *session.Session,
	bucket string,
	srcRootPath string, waitUploadWorker *sync.WaitGroup) {

	defer func() {
		waitUploadWorker.Done()
	}()

	uploader := s3manager.NewUploader(s3Session, func(u *s3manager.Uploader) {
		// 内存缓存，每个线程都会占有一份
		u.BufferProvider = s3manager.NewBufferedReadSeekerWriteToPool(5 * 1024 * 1024)
		u.PartSize = 10 * 1024 * 1024 // 100MB per part
		u.Concurrency = 1             //TODO: 并发
	})

	delTime := NewConstants().TaskParams.DelTime
	lastCreateTime := NewConstants().PathMTime[SYNC][srcRootPath]

	//统计扫描结果信息
	for file := range fileQueues {
		filename := filepath.Base(file.Path)
		if strings.HasPrefix(filename, ".wait-del-") || strings.HasPrefix(filename, "~$") {
			continue
		}

		statusFile := utils.IsStatusFileExist(file.Path)
		cTime := file.ChangeTime
		if cTime < lastCreateTime {
			if statusFile && delTime >= 0 && (delTime == 0 || cTime+delTime <= time.Now().Unix()) {
				// _ = os.Remove(file.Path)
				utils.RemoveFileStatusFile(file.Path)
			} //删除之前传完并满足删除条件的文件，版本1.9.2
		}
		//if cTime >= lastCreateTime { //，版本1.9.1
		if !statusFile { // 没有状态文件就要上传，不在验证最后修改时间，版本1.9.2
			fileChangeTime := time.Unix(cTime, 0)
			strChangeTime := fileChangeTime.Format("2006-01-02 15:04:05")

			err, dstPath := s3UploadFile(srcRootPath, bucket, file.Path, strChangeTime, uploader)
			syncFileRlt <- SyncFileRlt{
				Path:       file.Path,
				FileSize:   file.Size,
				Err:        err,
				DstPath:    dstPath,
				ChangeTime: strChangeTime,
			}

			if err == nil { //文件上传成功
				//log.Printf("file %s uploaded successfully, need delete %d, %d", file.Path, delTime, cTime)
				if delTime >= 0 && (delTime == 0 || cTime+delTime <= time.Now().Unix()) {
					//传完后，满足删除条件就删，主要解决磁盘快满和大量历史数据问题，版本1.9.1
					_ = os.Remove(file.Path) //直接删除的话不用生成状态文件
				} else {
					// 上传完成等待删除，生成一个等待删除的配置文件，命名为：.wait-del-filename.ini 版本1.9.2
					utils.CreateStatusFile(file.Path)
				}
			} else {
				log.Printf("file %s uploaded failed %s.", file.Path, err.Error())
			}
			//else { // 上传失败更新文件最后修改时间 版本1.9.1
			//	log.Printf("file %s uploaded failed %s.", file.Path, err.Error())
			//	_fileInfo, _err := os.Stat(file.Path)
			//	if _err == nil {
			//		os.Chtimes(file.Path, _fileInfo.ModTime(), time.Now())
			//	}
			//}
		}
	}
}

func formatFileSize(size int64) string {
	units := []string{"B", "KB", "MB", "GB", "TB", "PB", "EB", "ZB", "YB"}
	base := int64(1024)
	if size < base {
		return fmt.Sprintf("%d %s", size, units[0])
	}
	unitIndex := int(math.Log(float64(size)) / math.Log(float64(base)))
	formattedSize := float64(size) / math.Pow(float64(base), float64(unitIndex))
	return fmt.Sprintf("%.2f %s", formattedSize, units[unitIndex])
}

func S3UploadLogFile(logFilePath, bucket string, s3Session *session.Session) (string, error) {

	file, err := os.Open(logFilePath)
	if err != nil {
		log.Printf("Unable to open log file %q\n", err)
		return "", err
	}
	defer file.Close()

	uploader := s3manager.NewUploader(s3Session, func(u *s3manager.Uploader) {
		// Define a strategy that will buffer 25 MiB in memory
		u.BufferProvider = s3manager.NewBufferedReadSeekerWriteToPool(5 * 1024 * 1024)
		u.PartSize = 100 * 1024 * 1024 // 10MB per part
		u.Concurrency = 1              //并发
	})

	pathPrefix := NewConstants().TaskParams.PathPrefix
	if !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	currentDate := time.Now()
	dateString := currentDate.Format("2006-01-02")

	logKey := pathPrefix + "syncLogs/" + dateString + "/" + filepath.Base(logFilePath)
	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(logKey),
		Body:   file,
	})

	if err != nil {
		log.Printf("Unable to upload log %q to %q, %v\n", logFilePath, bucket, err)
		return "", err
	}

	return logKey, nil
}

func s3UploadFile(srcRootPath, bucket, srcPath, changeTime string, uploader *s3manager.Uploader) (error, string) {
	file, err := os.Open(srcPath)
	if err != nil {
		log.Printf("Unable to open file %q\n", err)
		return err, ""
	}
	defer file.Close()
	pathPrefix := NewConstants().TaskParams.PathPrefix

	rootPath := srcRootPath
	if utils.IsWindows() {
		rootPath = strings.Replace(srcRootPath, "\\", "/", -1)
		srcPath = strings.Replace(srcPath, "\\", "/", -1)
	}

	rootName := filepath.Base(rootPath)
	if strings.HasSuffix(rootPath, "/") && !strings.HasSuffix(pathPrefix, "/") {
		pathPrefix = pathPrefix + "/"
	}
	pathPrefix = pathPrefix + rootName + "/"
	dstPath := strings.Replace(srcPath, rootPath, pathPrefix, 1)

	// 基础标签
	fileTags := NewConstants().TaskParams.Tags
	myMetadata := map[string]*string{
		"Tag":      aws.String(url.QueryEscape(fileTags)),
		"Modified": aws.String(changeTime),
	}

	// 提取标签
	dstPath, _ = extractInfoFromFilename(dstPath, srcPath, &myMetadata)

	_, err = uploader.Upload(&s3manager.UploadInput{
		Bucket:   aws.String(bucket),
		Key:      aws.String(dstPath),
		Body:     file,
		Metadata: myMetadata,
	})

	return err, dstPath
}

// ExtractInfoFromFilename 根据提供的正则表达式和标签列表从文件名中提取信息
func extractInfoFromFilename(dstPath, srcPath string, myMetadata *map[string]*string) (string, error) {
	fetchField := NewConstants().TaskParams.FetchField
	pattern := NewConstants().TaskParams.FetchRegex
	tags := NewConstants().TaskParams.TagsKeyList
	tagsAsPath := NewConstants().TaskParams.TagsAsPath

	_newDstPath := dstPath

	// 从文件名中提取内容 开始
	_filename := filepath.Base(srcPath)

	if fetchField < 0 {
		return _newDstPath, fmt.Errorf("fetch info off")
	}

	fetchStr := srcPath
	if fetchField == 2 { // 2 从文件名中提取
		fetchStr = _filename
	}

	// 编译正则表达式
	re, err := regexp.Compile(pattern)
	if err != nil {
		return _newDstPath, fmt.Errorf("invalid regex pattern: %w", err)
	}

	// 执行匹配
	matches := re.FindStringSubmatch(fetchStr)
	if matches == nil {
		return _newDstPath, fmt.Errorf("no match found or invalid filename format")
	}

	// 创建结果映射
	newParentPath := ""
	for _, tag := range tags {
		// 跳过索引0，因为matches[0]是整个匹配的字符串
		if tag.Index < len(matches) {
			// result[tag.TagKey] = matches[tag.Index+1]
			_item := matches[tag.Index]
			(*myMetadata)[tag.TagKey] = aws.String(url.QueryEscape(_item))
			if newParentPath == "" {
				newParentPath = _item
			} else {
				newParentPath = newParentPath + "/" + _item
			}
		}
	}

	if fetchField == 2 && tagsAsPath == 1 {
		newParentPath = newParentPath + "/" + _filename
		log.Printf("Debug  newParentPath=%s\n", newParentPath)
		return strings.Replace(dstPath, fetchStr, newParentPath, 1), nil
	} else {
		return _newDstPath, nil
	}
}
