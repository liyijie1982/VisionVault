package service

import (
	"SyncAgent/utils"
	"log"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

// scanDirAndSendTask 递归扫描指定目录,将符合条件的文件信息发送到任务队列
// srcRootPath: 源根目录路径
// currentPath: 当前正在扫描的目录路径
// fileQueues: 用于接收文件信息的channel
// 返回值: 读取目录过程中的错误,如果成功则返回nil
func scanDirAndSendTask(srcRootPath, currentPath string, fileQueues chan<- SyncFileInfo) error {
	infos, err := utils.ReadSubDir(currentPath)
	if err != nil {
		log.Println(err)
		return err
	}

	for _, info := range infos {
		_filePath := filepath.Join(currentPath, info.Name())
		if info.IsDir() {
			_ = scanDirAndSendTask(srcRootPath, _filePath, fileQueues)
		} else {
			if !NewConstants().HasFilter(_filePath) {
				//sys := info.Sys()
				//stat := sys.(*syscall.Stat_t)
				fileQueues <- SyncFileInfo{
					Path:       _filePath,
					Size:       info.Size(),
					ChangeTime: info.ModTime().Unix(), //stat.Ctim.Sec,
				}
			}
		}
	}
	return nil
}

// ScanFileAndSendTask 扫描指定目录,将符合条件的文件信息发送到任务队列
// srcRootPath: 源根目录路径
// currentPath: 当前正在扫描的目录路径
// fileQueues: 用于接收文件信息的channel
// 返回值: 读取目录过程中的错误,如果成功则返回nil
func ScanFileAndSendTask(srcRootPath, currentPath string, fileQueues chan<- SyncFileInfo) {
	defer func() {
		close(fileQueues)
	}()

	_ = scanDirAndSendTask(srcRootPath, currentPath, fileQueues)

	log.Println("ScanFileAndSendTask finished.")
}

// ReadLocalFolder 读取本地文件夹信息
// path: 本地文件夹路径
// 返回值: 本地文件夹信息,如果成功则返回nil
func ReadLocalFolder(path string) ([]utils.Storage, error) {

	if path == "/" && utils.IsWindows() { //windows root
		return utils.GetStorageInfo(), nil
	} else {
		if !utils.IsWindows() && path != "" && !strings.HasPrefix(path, "/") {
			path = "/" + path
		}
		dir, err := utils.ReadSubDir(path)
		if err != nil {
			return nil, err
		} else {
			var localFiles []utils.Storage
			for _, info := range dir {
				if !info.IsDir() {
					continue
				}
				_filepath := filepath.Join(path, info.Name())
				localFiles = append(localFiles, utils.Storage{Name: _filepath})
			}
			return localFiles, nil
		}
	}
}

func LocalUploadWorker(fileQueues chan SyncFileInfo,
	syncFileRlt chan SyncFileRlt,
	localDstRootPath string,
	srcRootPath string, waitUploadWorker *sync.WaitGroup) {

	defer func() {
		waitUploadWorker.Done()
	}()

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
				utils.RemoveFileStatusFile(file.Path)
			} //删除之前传完并满足删除条件的文件，版本1.9.2
		}

		if !statusFile { // 没有状态文件就要上传，不在验证最后修改时间，版本1.9.2
			fileChangeTime := time.Unix(cTime, 0)
			strChangeTime := fileChangeTime.Format("2006-01-02 15:04:05")

			err, dstPath := LocalUploadFile(srcRootPath, localDstRootPath, file.Path, strChangeTime)
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
		}
	}
}

func LocalUploadFile(srcRootPath, localDstRootPath, srcPath, changeTime string) (error, string) {
	pathPrefix := NewConstants().TaskParams.PathPrefix

	rootPath := srcRootPath
	if utils.IsWindows() {
		rootPath = utils.FormatPathToWindowsSeparator(srcRootPath)
		srcPath = utils.FormatPathToWindowsSeparator(srcPath)
		pathPrefix = utils.FormatPathToWindowsSeparator(pathPrefix)
	}

	rootName := filepath.Base(rootPath)
	if strings.HasSuffix(rootName, ":") { //直接上传 Windows 磁盘根目录时，要去掉盘符
		rootName = ""
	}

	utils.AddSeparator(rootPath)
	if len(pathPrefix) > 0 {
		utils.AddSeparator(pathPrefix)
	}
	dstRootPath := utils.AddSeparator(localDstRootPath)

	pathPrefix = utils.AddSeparator(dstRootPath + pathPrefix + rootName)
	dstPath := strings.Replace(srcPath, rootPath, pathPrefix, 1)

	// 确保目标目录存在
	dstDir := filepath.Dir(dstPath)
	err := os.MkdirAll(dstDir, 0755)
	if err != nil {
		log.Printf("Unable to create directory %q: %v\n", dstDir, err)
		return err, ""
	}

	// 复制文件
	err = utils.CopyFile(srcPath, dstPath)
	if err != nil {
		log.Printf("Unable to copy file from %s to %s: %v\n", srcPath, dstPath, err)
		return err, ""
	}

	return nil, dstPath
}
