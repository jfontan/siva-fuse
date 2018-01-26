
# siva-fuse

Fuse filesystem for [siva](https://github.com/src-d/go-siva#%C5%9Biva-format-%E0%A4%B6%E0%A4%BF%E0%A4%B5---) files.

This filesystem mounts a local directory with siva files inside and you'll be able to access without need to unpack them. The siva files will appear as directories and can be accessed with standard posix commands. The mounted filesystem is read only.

**Warning**: This project is still an experiment without proper testing. Use at your own risk.

## Installation

With a working go language installation you can install the command with:

``` 
go get github.com/jfontan/siva-fuse
```

## Usage

Mount the directory `/storage/siva` in `/mnt`:

```
siva-fuse /storage/siva /mnt
```

Umount filesystem:

```
fusermount -u /mnt
```

It should be compatible with MacOS X https://osxfuse.github.io/
