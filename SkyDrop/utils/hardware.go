package utils

/**
该方法只支持Windows编译
*/
import (
	"log"
	"time"

	//"github.com/StackExchange/wmi"
	"github.com/shirou/gopsutil/v3/cpu"
	"github.com/shirou/gopsutil/v3/disk"
	"github.com/shirou/gopsutil/v3/mem"
)

type Storage struct {
	Name       string `json:"name"`
	FileSystem string `json:"fileSystem"`
	Total      uint64 `json:"total"`
	Free       uint64 `json:"free"`
}

type storageInfo struct {
	Name       string
	Size       uint64
	FreeSpace  uint64
	FileSystem string
}

// func GetStorageInfo() []Storage {
// 	var storageInfo []storageInfo
// 	var localStorages []Storage
// 	err := wmi.Query("Select * from Win32_LogicalDisk", &storageInfo)
// 	if err != nil {
// 		log.Println(err.Error())
// 		return localStorages
// 	}

//		for _, storage := range storageInfo {
//			info := Storage{
//				Name:       storage.Name,
//				FileSystem: storage.FileSystem,
//				Total:      storage.Size,
//				Free:       storage.FreeSpace,
//			}
//			localStorages = append(localStorages, info)
//		}
//		log.Println(localStorages)
//		return localStorages
//	}
func GetStorageInfo() []Storage {
	var localStorages []Storage
	devices, err := disk.Partitions(true) //所有分区
	if err != nil {
		log.Panicln(err.Error())
		return localStorages
	}

	for _, device := range devices {
		usageStat, err := disk.Usage(device.Device)
		if err == nil {
			info := Storage{
				Name:       device.Device,
				FileSystem: device.Fstype,
				Total:      usageStat.Total,
				Free:       usageStat.Free,
			}
			localStorages = append(localStorages, info)
		}
	}

	//log.Println(localStorages)
	return localStorages
}

type ComputerMonitor struct {
	CPU float64 `json:"cpu"`
	Mem float64 `json:"mem"`
}

// GetCPUPercent 获取CPU使用率
func GetCPUPercent() float64 {
	percent, err := cpu.Percent(10*time.Second, false)
	if err != nil {
		log.Println(err.Error())
		return -1
	}
	return percent[0]
}

// GetMemPercent 获取内存使用率
func GetMemPercent() float64 {
	memInfo, err := mem.VirtualMemory()
	if err != nil {
		log.Println(err.Error())
		return -1
	}
	return memInfo.UsedPercent
}

func GetCpuMem() ComputerMonitor {
	var res ComputerMonitor
	res.CPU = GetCPUPercent()
	res.Mem = GetMemPercent()
	//fmt.Printf("%v", res)
	//log.Printf("CPU使用率：%.2f%%\n", res.CPU)
	//log.Printf("内存使用率：%.2f%%\n", res.Mem)
	return res
}
