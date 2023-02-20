package utils

import (
	"fmt"
	"net"
)

func GetOrSelectIPv4Addr() []string {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		fmt.Println(err)
	}

	//ipList := make([]net.IP, 0)

	var ipAddress = make([]string, 0)
	for _, address := range addrs {
		if ipnet, ok := address.(*net.IPNet); ok && !ipnet.IP.IsLoopback() && !ipnet.IP.IsLinkLocalUnicast() {
			if ipnet.IP.To4() != nil {
				//ipList = append(ipList, ipnet.IP.To4())
				ipAddress = append(ipAddress, ipnet.IP.To4().String())
			}
		}
	}

	return ipAddress
}
