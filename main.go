// Copyright 2012 Google Inc. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package main

import (
	"flag"
	"log"
	"os"

	"github.com/hanwen/go-fuse/fuse"
	"github.com/hanwen/go-mtpfs/fs"
	"github.com/hanwen/go-mtpfs/mtp"
)

func main() {
	fsdebug := flag.Bool("fs-debug", false, "switch on FS debugging")
	mtpDebug := flag.Bool("mtp-debug", false, "switch on MTP debugging")
	dataDebug := flag.Bool("data-debug", false, "switch on data debugging")
	usbTimeout := flag.Int("usb-timeout", 2000, "timeout in milliseconds")
	vfat := flag.Bool("vfat", true, "assume removable RAM media uses VFAT, and rewrite names.")
	other := flag.Bool("allow-other", false, "allow other users to access mounted fuse. Default: false.")
	deviceFilter := flag.String("dev", "", "regular expression to filter devices.")
	storageFilter := flag.String("storage", "", "regular expression to filter storage areas.")
	android := flag.Bool("android", true, "use android extensions if available")
	flag.Parse()

	if len(flag.Args()) != 1 {
		log.Fatalf("Usage: %s [options] MOUNT-POINT\n", os.Args[0])
	}
	mountpoint := flag.Arg(0)

	dev, err := mtp.SelectDevice(*deviceFilter)
	if err != nil {
		log.Fatalf("detect failed: %v", err)
	}
	defer dev.Close()

	dev.Timeout = *usbTimeout
	if err = dev.Configure(); err != nil {
		log.Fatalf("Configure failed: %v", err)
	}

	sids, err := fs.SelectStorages(dev, *storageFilter)
	if err != nil {
		log.Fatalf("selectStorages failed: %v", err)
	}

	dev.DebugPrint = *mtpDebug
	dev.DataPrint = *dataDebug
	opts := fs.DeviceFsOptions{
		RemovableVFat: *vfat,
		Android: *android,
	}
	fs, err := fs.NewDeviceFs(dev, sids, opts)
	if err != nil {
		log.Fatalf("NewDeviceFs failed: %v", err)
	}
	conn := fuse.NewFileSystemConnector(fs, fuse.NewFileSystemOptions())
	rawFs := fuse.NewLockingRawFileSystem(conn)

	mount := fuse.NewMountState(rawFs)
	mOpts := &fuse.MountOptions{
		AllowOther: *other,
	}
	if err := mount.Mount(mountpoint, mOpts); err != nil {
		log.Fatalf("mount failed: %v", err)
	}

	conn.Debug = *fsdebug
	mount.Debug = *fsdebug
	log.Printf("starting FUSE %v", fuse.Version())
	mount.Loop()
	fs.OnUnmount()
}
