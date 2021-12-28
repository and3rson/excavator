package main

import (
	"fmt"
	"net"

	log "github.com/sirupsen/logrus"
)

func ResolveAddr(hostname string) (*net.SRV, *net.TCPAddr, error) {
	cname, addrs, err := net.LookupSRV("minecraft", "tcp", hostname)
	var service *net.SRV
	if err == nil {
		log.Debugf("resolve: cname: %s, addrs: %v", cname, addrs)
		service = addrs[0]
		log.Infof("resolve: target/port: %s:%d", service.Target, service.Port)
	} else {
		service = &net.SRV{
			Target: hostname,
			Port:   25565,
		}
		log.Warnf("resolve: cname lookup failed, using hostname directly")
	}

	addr, err := net.ResolveTCPAddr("tcp", service.Target+":"+fmt.Sprint(service.Port))
	if err != nil {
		return nil, nil, err
	}

	return service, addr, nil
}
