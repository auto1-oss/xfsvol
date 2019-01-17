package main

import (
	"log"
	"sync"

	"github.com/cirocosta/xfsvol/manager"
	"github.com/pkg/errors"

	v "github.com/docker/go-plugins-helpers/volume"
)

type DriverConfig struct {
	HostMountpoint string
	DefaultSize    string
}

type Driver struct {
	defaultSize string
	manager     *manager.Manager
	sync.Mutex
}

func NewDriver(cfg DriverConfig) (d Driver, err error) {
	if cfg.HostMountpoint == "" {
		err = errors.Errorf("HostMountpoint must be specified")
		return
	}

	if cfg.DefaultSize == "" {
		err = errors.Errorf("DefaultSize must be specified")
		return
	}

	m, err := manager.New(manager.Config{
		Root: cfg.HostMountpoint,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"Couldn't initiate fs manager mounting at %s",
			cfg.HostMountpoint)
		return
	}

	d.defaultSize = cfg.DefaultSize
	log.Println("driver initiated")
	d.manager = &m

	return
}

func (d Driver) Create(req *v.CreateRequest) (err error) {

	size, present := req.Options["size"]
	if !present {
		log.Printf("%s no size opt found, using default", req.Name)
		size = d.defaultSize
	}

	sizeInBytes, err := manager.FromHumanSize(size)
	if err != nil {
		err = errors.Errorf(
			"couldn't convert specified size [%s] into bytes",
			size)
		return
	}

	d.Lock()
	defer d.Unlock()

	log.Printf("%s starting creation", req.Name)

	absHostPath, err := d.manager.Create(manager.Volume{
		Name: req.Name,
		Size: sizeInBytes,
	})
	if err != nil {
		err = errors.Wrapf(err,
			"manager failed to create volume %s",
			req.Name)
		return
	}

	log.Printf("abs-host-path: %s finished creating volume", absHostPath)
	return
}

func (d Driver) List() (resp *v.ListResponse, err error) {

	d.Lock()
	defer d.Unlock()

	log.Println("starting volume listing")

	vols, err := d.manager.List()
	if err != nil {
		err = errors.Wrapf(err,
			"manager failed to list volumes")
		return
	}

	resp = new(v.ListResponse)
	resp.Volumes = make([]*v.Volume, len(vols))
	for idx, vol := range vols {
		resp.Volumes[idx] = &v.Volume{
			Name: vol.Name,
		}
	}

	log.Printf("number-of-volumes %s", len(vols))
	return
}

func (d Driver) Get(req *v.GetRequest) (resp *v.GetResponse, err error) {

	d.Lock()
	defer d.Unlock()

	vol, found, err := d.manager.Get(req.Name)
	if err != nil {
		err = errors.Wrapf(err,
			"manager failed to get volume named %s",
			req.Name)
		return
	}

	if !found {
		err = errors.Errorf("volume %s not found", req.Name)
		return
	}

	resp = new(v.GetResponse)
	resp.Volume = &v.Volume{
		Name:       req.Name,
		Mountpoint: vol.Path,
	}

	log.Printf("finished retrieving volume %s", req.Name)
	return
}

func (d Driver) Remove(req *v.RemoveRequest) (err error) {

	d.Lock()
	defer d.Unlock()

	err = d.manager.Delete(req.Name)
	if err != nil {
		err = errors.Wrapf(err,
			"manager failed to delete volume named %s",
			req.Name)
		return
	}
	log.Printf("volume %s removed", req.Name)
	return
}

func (d Driver) Path(req *v.PathRequest) (resp *v.PathResponse, err error) {

	d.Lock()
	defer d.Unlock()

	vol, found, err := d.manager.Get(req.Name)
	if err != nil {
		err = errors.Wrapf(err,
			"manager failed to retrieve volume named %s",
			req.Name)
		return
	}

	if !found {
		err = errors.Errorf("volume %s not found", req.Name)
		return
	}
	log.Printf("path %s retrieved for volume %s", vol.Path, req.Name)

	resp = new(v.PathResponse)
	resp.Mountpoint = vol.Path
	return
}

func (d Driver) Mount(req *v.MountRequest) (resp *v.MountResponse, err error) {

	d.Lock()
	defer d.Unlock()

	vol, found, err := d.manager.Get(req.Name)
	if err != nil {
		err = errors.Wrapf(err,
			"failed to retrieve volume named %s",
			req.Name)
		return
	}

	if !found {
		err = errors.Errorf("volume %s not found", req.Name)
		return
	}

	log.Printf("finished mounting volume %s", req.Name)

	resp = new(v.MountResponse)
	resp.Mountpoint = vol.Path
	return
}

func (d Driver) Unmount(req *v.UnmountRequest) (err error) {

	d.Lock()
	defer d.Unlock()

	log.Printf("finished unmounting %s", req.Name)

	return
}

// TODO is it global?
func (d Driver) Capabilities() (resp *v.CapabilitiesResponse) {
	resp = &v.CapabilitiesResponse{
		Capabilities: v.Capability{
			Scope: "global",
		},
	}
	return
}
