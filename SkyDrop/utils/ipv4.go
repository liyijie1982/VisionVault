package utils

import (
	"fmt"
	"net"
	"os"
	"strings"
)

func GetLocalHostName() string {
	name, err := os.Hostname()
	if err != nil {
		fmt.Printf("fail to get net hostname: %v\n", err)
	}
	return name
}

// shortenPrefix 函数会返回给定IP前缀的上一级子网前缀。
// 例如，输入 "192.168.1" 返回 "192.168"。
func shortenPrefix(ipPrefix string) string {
	parts := strings.Split(ipPrefix, ".")
	if len(parts) <= 1 {
		return ""
	}
	return strings.Join(parts[:len(parts)-1], ".")
}

func GetOneLocalIp(ipPrefix string) string {
	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		return ""
	}

	var fallbackIP string // 用于存储第一个非127.0.0.1的IP作为备选

	// 如果ipPrefix为空，直接返回第一个非127.0.0.1的IP
	if ipPrefix == "" {
		for _, address := range interfaceAddr {
			if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
				if ip := ipNet.IP.To4(); ip != nil {
					return ip.String()
				}
			}
		}
		return ""
	}

	// 解析多段IP前缀（用逗号分隔）
	prefixes := strings.Split(ipPrefix, ",")
	for i := range prefixes {
		prefixes[i] = strings.TrimSpace(prefixes[i])
	}

	// 遍历所有网络接口
	for _, address := range interfaceAddr {
		if ipNet, ok := address.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ip := ipNet.IP.To4(); ip != nil {
				ipStr := ip.String()

				// 记录第一个非127.0.0.1的IP作为备选
				if fallbackIP == "" {
					fallbackIP = ipStr
				}

				// 检查是否匹配任何一个前缀
				for _, prefix := range prefixes {
					if strings.HasPrefix(ipStr, prefix) {
						return ipStr
					}
				}
			}
		}
	}

	// 如果没有找到匹配的IP，返回第一个非127.0.0.1的IP
	return fallbackIP
}

func GetLocalIp(ipRange string) (ip string) {

	interfaceAddr, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Printf("fail to get net interfaces ipAddress: %v\n", err)
		return ip
	}

	for _, address := range interfaceAddr {
		ipNet, isVailIpNet := address.(*net.IPNet)
		if isVailIpNet && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				ip = ipNet.IP.String()
				if strings.HasPrefix(ip, ipRange) {
					return ip
				}
			}
		}
	}
	return ipRange
}

func GetLocalMac() (mac string) {
	// 获取本机的MAC地址
	interfaces, err := net.Interfaces()
	if err != nil {
		fmt.Println("Poor soul, here is what you got:", err.Error())
	}
	for _, inter := range interfaces {
		fmt.Println(inter.Name)
		mac := inter.HardwareAddr //获取本机MAC地址
		fmt.Println("MAC ===== ", mac)
	}
	fmt.Println("MAC = ", mac)
	return mac
}
