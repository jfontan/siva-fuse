package main

import (
	"os"

	"github.com/jfontan/siva-fuse/sivafuse"

	"github.com/hanwen/go-fuse/fuse/nodefs"
	"github.com/hanwen/go-fuse/fuse/pathfs"
)

func printHelp() {
	println("You have to provide <siva dir> <mount point>")
}

// NewSivaNodeFs creates a new PathNodeFs from the provided path
func NewSivaNodeFs(sivaDir string) *pathfs.PathNodeFs {
	root := sivafuse.NewRootSivaFs(sivaDir)

	pathOpts := &pathfs.PathNodeFsOptions{}
	rootfs := pathfs.NewPathNodeFs(root, pathOpts)

	return rootfs
}

func main() {
	if len(os.Args) != 3 {
		printHelp()
		os.Exit(1)
	}

	sivaDir := os.Args[1]
	mountDir := os.Args[2]

	rootfs := NewSivaNodeFs(sivaDir)

	opts := nodefs.NewOptions()
	opts.Debug = false

	state, _, err := nodefs.MountRoot(mountDir, rootfs.Root(), opts)
	if err != nil {
		panic(err)
	}

	state.Serve()
}
