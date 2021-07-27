package main

import (
	"fmt"
	"log"
	"time"

	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"

	"github.com/vmware/vmw-guestinfo/rpcvmx"
	"github.com/vmware/vmw-guestinfo/vmcheck"
)

func main() {
	log.Println("Starting up config.")
	time.Sleep(15 * time.Second)

	// Evaluate whether the workload is running on vSphere
	log.Println("Checking if VM or not.")
	isVM, err := vmcheck.IsVirtualWorld()
	if err != nil {
		log.Fatalf("Error: %s", err)
	}

	if !isVM {
		log.Fatalf("ERROR: not in a virtual world.")
	}
	log.Println("Running on vSphere.")

	vsphereConfig := rpcvmx.NewConfig()

	ipAddress := ""
	if out, err := vsphereConfig.String("guestinfo.ipxe.net0.ip", ""); err != nil {
		log.Fatalf("ERROR: String failed with %s", err)
	} else {
		fmt.Printf("%s\n", out)
		ipAddress = out
	}

	eth0, _ := netlink.LinkByName("eth0")
	addr, _ := netlink.ParseAddr(ipAddress + "/24")

	// remove existing addresses from link
	addresses, err := netlink.AddrList(eth0, nl.FAMILY_ALL)
	if err != nil {
		fmt.Errorf("Cannot list addresses on interface: %v", err)
	}
	for _, addr := range addresses {
		if err := netlink.AddrDel(eth0, &addr); err != nil {
			fmt.Errorf("Cannot remove address from interface: %v", err)
		}
	}
	netlink.AddrReplace(eth0, addr)
	time.Sleep(600 * time.Second)
}
