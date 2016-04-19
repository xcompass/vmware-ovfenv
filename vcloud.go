package main

/*************************************
  Example of vCloud properties
# /usr/share/oem/bin/vmtoolsd --cmd "info-get guestinfo.ovfenv"

<?xml version="1.0" encoding="UTF-8"?>
<Environment
     xmlns="http://schemas.dmtf.org/ovf/environment/1"
     xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
     xmlns:oe="http://schemas.dmtf.org/ovf/environment/1"
     xmlns:ve="http://www.vmware.com/schema/ovfenv"
     oe:id=""
     ve:vCenterId="vm-108831">
   <PlatformSection>
      <Kind>VMware ESXi</Kind>
      <Version>5.5.0</Version>
      <Vendor>VMware, Inc.</Vendor>
      <Locale>en</Locale>
   </PlatformSection>
   <PropertySection>
         <Property oe:key="guestinfo.coreos.config.data" oe:value=""/>
         <Property oe:key="guestinfo.coreos.config.data.encoding" oe:value=""/>
         <Property oe:key="guestinfo.coreos.config.url" oe:value=""/>
         <Property oe:key="guestinfo.dns.server.0" oe:value=""/>
         <Property oe:key="guestinfo.dns.server.1" oe:value=""/>
         <Property oe:key="guestinfo.hostname" oe:value=""/>
         <Property oe:key="guestinfo.interface.0.dhcp" oe:value="yes"/>
         <Property oe:key="guestinfo.interface.0.ip.0.address" oe:value=""/>
         <Property oe:key="guestinfo.interface.0.mac" oe:value=""/>
         <Property oe:key="guestinfo.interface.0.name" oe:value=""/>
         <Property oe:key="guestinfo.interface.0.role" oe:value="public"/>
         <Property oe:key="guestinfo.interface.0.route.0.destination" oe:value=""/>
         <Property oe:key="guestinfo.interface.0.route.0.gateway" oe:value=""/>
         <Property oe:key="vCloud_UseSysPrep" oe:value="None"/>
         <Property oe:key="vCloud_bitMask" oe:value="1"/>
         <Property oe:key="vCloud_bootproto_0" oe:value="static"/>
         <Property oe:key="vCloud_computerName" oe:value="myvm"/>
         <Property oe:key="vCloud_dns1_0" oe:value="8.8.8.8"/>
         <Property oe:key="vCloud_dns2_0" oe:value="4.4.4.4"/>
         <Property oe:key="vCloud_gateway_0" oe:value="10.93.1.254"/>
         <Property oe:key="vCloud_ip_0" oe:value="10.93.1.23"/>
         <Property oe:key="vCloud_macaddr_0" oe:value="00:51:52:03:0e:42"/>
         <Property oe:key="vCloud_markerid" oe:value="a411b0e0-6515-4e63-b6ac-712c51752af8"/>
         <Property oe:key="vCloud_netmask_0" oe:value="255.255.255.0"/>
         <Property oe:key="vCloud_numnics" oe:value="1"/>
         <Property oe:key="vCloud_primaryNic" oe:value="0"/>
         <Property oe:key="vCloud_reconfigToken" oe:value="550746732"/>
         <Property oe:key="vCloud_resetPassword" oe:value="0"/>
         <Property oe:key="vCloud_suffix_0" oe:value=""/>
   </PropertySection>
   <ve:EthernetAdapterSection>
      <ve:Adapter ve:mac="00:51:52:03:0e:42" ve:network="MYNETWORK" ve:unitNumber="7"/>
   </ve:EthernetAdapterSection>
</Environment>

**************************************/

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"strings"
	"text/template"

	"github.com/sigma/vmw-guestinfo/rpcvmx"
	"github.com/sigma/vmw-guestinfo/vmcheck"
	"github.com/sigma/vmw-ovflib"
)

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func readConfig(key string) (string, error) {
	data, err := rpcvmx.NewConfig().String(key, "")
	if err == nil {
		log.Printf("Read from %q: %q\n", key, data)
	} else {
		log.Printf("Failed to read from %q: %v\n", key, err)
	}
	return data, err
}

func checkCustomizationParameters(props map[string]string) bool {
	// if customization is enabled, vCloud_* properties will be populated
	for key, _ := range props {
		if strings.HasPrefix(key, "vCloud_") {
			return true
		}
	}
	return false
}

func buildNetworkUnit(props map[string]string) {
	ipNet := net.IPNet{
		IP:   net.ParseIP(props["vCloud_ip_0"]),
		Mask: net.IPMask(net.ParseIP(props["vCloud_netmask_0"])),
	}
	props["ip"] = ipNet.String()
	tpl, err := template.New("unit").Parse(`
[Match]
Name=ens192

[Network]
Address={{ .ip }}
Gateway={{ .vCloud_gateway_0 }}
DNS={{ .vCloud_dns1_0 }}
DNS={{ .vCloud_dns2_0 }}
`)
	check(err)

	// create file
	f, err := os.Create("/etc/systemd/network/static.network")
	defer f.Close()

	err = tpl.Execute(f, props)
	check(err)

	log.Printf("static.network is generated.")
}

func setHostname(hostname string) {
	err := exec.Command("hostnamectl", "set-hostname", hostname).Run()
	check(err)
}

func main() {
	if !vmcheck.IsVirtualWorld() {
		fmt.Println("not in a virtual world... :(")
		return
	}

	data, err := readConfig("ovfenv")
	check(err)
	if data == "" {
		log.Printf("No data from ovfenv")
		return
	}

	log.Printf("Using OVF environment from guestinfo\n")
	env := ovf.ReadEnvironment([]byte(data))
	props := env.Properties
	if !checkCustomizationParameters(props) {
		log.Printf("No customization parameters. Skipping...")
		return
	}
	buildNetworkUnit(props)

	hostname := props["vCloud_computerName"]
	setHostname(hostname)
}
