package ip

import (
	"fmt"
	"net"
	"os"
)

func LocalIp() (ip string, err error) {
	address, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
	for _, info := range address {
		// 检查ip地址判断是否回环地址
		if ipNet, ok := info.(*net.IPNet); ok && !ipNet.IP.IsLoopback() {
			if ipNet.IP.To4() != nil {
				return ipNet.IP.String(), nil
			}
			return "", fmt.Errorf("error")
		}
		return "", fmt.Errorf("error")
	}
	return "", fmt.Errorf("error")
}
