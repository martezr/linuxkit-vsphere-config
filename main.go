package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"strings"
	"time"

	"github.com/martezr/linuxkit-vsphere-config/vip"
	"github.com/vishvananda/netlink"
	"github.com/vishvananda/netlink/nl"

	"github.com/vmware/vmw-guestinfo/rpcvmx"
	"github.com/vmware/vmw-guestinfo/vmcheck"
)

func main() {
	log.Println("Starting up config.")

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
	err = netlink.LinkSetUp(eth0)
	if err != nil {
		log.Fatalf("error bringing up the link: %v", err)
	}
	log.Println("Brought up NIC.")
	time.Sleep(20 * time.Second)
	createVIP()
}

func createVIP() {

	id := ""
	bind := "0.0.0.0"
	peersList := ""
	networkInterface := "eth0"
	virtualIP := "10.0.0.3"

	netlinkNetworkConfigurator, error := vip.NewNetlinkNetworkConfigurator(virtualIP, networkInterface)
	if error != nil {
		os.Exit(-1)
	}

	peers := vip.Peers{}

	if len(peersList) > 0 {
		for _, peer := range strings.Split(peersList, ",") {
			peerTokens := strings.Split(peer, "=")

			if len(peerTokens) != 2 {
				os.Exit(-1)
			}

			peers[peerTokens[0]] = peerTokens[1]
		}
	}

	logger := vip.Logger{}

	vipManager := vip.NewVIPManager(id, bind, peers, logger, netlinkNetworkConfigurator)
	if error := vipManager.Start(); error != nil {
		os.Exit(-1)
	}

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, os.Interrupt)

	<-signalChan

	vipManager.Stop()
}
