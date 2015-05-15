package main

import (
	"fmt"

	"github.com/coreos/coreos-cloudinit/Godeps/_workspace/src/github.com/sigma/vmw-guestinfo/rpcvmx"
	"github.com/coreos/coreos-cloudinit/Godeps/_workspace/src/github.com/sigma/vmw-guestinfo/vmcheck"
)

func main() {
	if !vmcheck.IsVirtualWorld() {
		fmt.Println("not in a virtual world... :(")
		return
	}

	version, typ := vmcheck.GetVersion()
	fmt.Println(version, typ)

	config := rpcvmx.NewConfig()
	fmt.Println(config.GetString("foo", "foo"))
	fmt.Println(config.GetString("bar", "foo"))
}
