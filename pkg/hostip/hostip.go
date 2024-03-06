// Package hostip provides function to detect host IP.
//
// В постановке задачи не сказано откуда агенту брать IP хоста.
//
// Личное неопытное мнение касательно задачи из урока - затея в целом странная.
// Считаю, что ip уместнее передавать через конфиг. Протестировать на реальном
// продакшн сервере возможности не было. А на домашнем (локальном) ПК не имеет
// смысла. Дергать сторонние сервисы как-то не хочется тоже.
//
// - Перебор интерфейсов net.Interfaces() может приводить к неправильным
// результатам.
// - В другом способе, `net.Dial().LocalAddr().(*net.UDPAddr).IP`, не уверен,
// проверить не было возможности, т.к. в распоряжении нет боевого VPS/VDS.
package hostip

import (
	"net"
)

// FindHostIP tries to detect current host IP.
// Returns empty string when nothing found. And non-nil error
// when smth went wrong.
func FindHostIP() (string, error) {
	ipNet, err := getOutboundIP()
	if err != nil {
		return "", err
	}

	return ipNet.String(), nil
}

// getOutboundIP - get preferred outbound ip of this machine.
//
// source: https://stackoverflow.com/a/37382208/4929867
func getOutboundIP() (ip net.IP, err error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP, nil
}

// method2 - способ дает много неправильных ip.
// На машине может быть множество сетевых интерфейсов,
// каждый со своим ip, что не помогает задаче.
//
// source: https://stackoverflow.com/a/23558495/4929867
// func method2() {
// 	ifaces, err := net.Interfaces()
// 	if err != nil {
// 		panic(err)
// 	}
// 	for _, i := range ifaces {
// 		addrs, err := i.Addrs()
// 		if err != nil {
// 			panic(err)
// 		}
// 		for _, addr := range addrs {
// 			var ip net.IP
// 			switch v := addr.(type) {
// 			case *net.IPNet:
// 				ip = v.IP
// 			case *net.IPAddr:
// 				ip = v.IP
// 			}
// 			fmt.Printf("ip.IsLoopback(): %t,\tip: %v\n", ip.IsLoopback(), ip)
// 		}
// 	}
// }
