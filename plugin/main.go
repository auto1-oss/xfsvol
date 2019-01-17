package main

import (
	"log"

	"github.com/alexflint/go-arg"

	v "github.com/docker/go-plugins-helpers/volume"
)

const (
	socketAddress = "/run/docker/plugins/xfsvol.sock"
)

type config struct {
	HostMountpoint string `arg:"--host-mountpoint,env:HOST_MOUNTPOINT,help:xfs-mounted filesystem to create volumes"`
	DefaultSize    string `arg:"--default-size,env:DEFAULT_SIZE,help:default size to use as quota"`
	Debug          bool   `arg:"env:DEBUG,help:enable debug logs"`
}

var (
	version string = "1.0.4"
	args           = &config{
		HostMountpoint: "/mnt/xfs/volumes",
		DefaultSize:    "512M",
		Debug:          false,
	}
)

func main() {
	arg.MustParse(args)

	log.Printf("version: %s, socket-address: %s, initializing plugin", version, socketAddress)

	d, err := NewDriver(DriverConfig{
		HostMountpoint: args.HostMountpoint,
		DefaultSize:    args.DefaultSize,
	})
	if err != nil {
		log.Fatalf("%s failed to initialize XFS volume driver", err)
	}

	h := v.NewHandler(d)
	err = h.ServeUnix(socketAddress, 0)
	if err != nil {
		log.Fatalf("%s failed to server volume plugin api over unix socket %s", err, socketAddress)
	}

	return
}
