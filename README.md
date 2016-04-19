VMWare vCloud Guest Customization Tool for CoreOS
-------------------------------------------------

This tools has basic support to customize CoreOS in vCloud. Since CoreOS doesn't support vCloud. This tool can be used to setup static network and hostname, which are provided by vCloud in OVF environment.

# Compile
```bash
cd $GOPATH
go install src/github.com/xcompass/vmware-ovfenv/vcloud.go
```

# Run
```bash
bin/vcloud
```
It will read information set up by vCloud from OVF environment and write `static.network` to `/etc/systemd/network`. It will also use `hostnamectl` to setup hostname with value in `vCloud_computerName`.
