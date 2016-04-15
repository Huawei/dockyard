package rados

import (
	"encoding/binary"
	"fmt"
	"io"
	"os"
	"path"
	"strconv"

	"github.com/ceph/go-ceph/rados"
	"github.com/satori/go.uuid"

	"github.com/containerops/dockyard/backend/factory"
	"github.com/containerops/dockyard/utils"
	"github.com/containerops/dockyard/utils/setting"
)

// Prefix all the stored blob
const objectBlobPrefix = "blob:"

// Stripes objects size to 4M
const defaultChunkSize = 4 << 20
const defaultXattrTotalSizeName = "total-size"

func init() {
	factory.Register("rados", &radosdesc{})
}

type radosdesc struct {
	Conn      *rados.Conn
	Ioctx     *rados.IOContext
	Chunksize uint64
}

func (d *radosdesc) New() (factory.DrvInterface, error) {
	chunksize := uint64(defaultChunkSize)
	if setting.Chunksize != "" {
		if tmp, err := strconv.Atoi(setting.Chunksize); err != nil {
			return nil, fmt.Errorf("The chunksize parameter should be a number")
		} else {
			chunksize = uint64(tmp)
		}
	}

	var err error
	var conn *rados.Conn
	if setting.Username != "" {
		conn, err = rados.NewConnWithUser(setting.Username)
	} else {
		conn, err = rados.NewConn()
	}
	if err != nil {
		return nil, err
	}

	if err := conn.ReadDefaultConfigFile(); err != nil {
		return nil, err
	}

	if err := conn.Connect(); err != nil {
		return nil, err
	}

	ioctx, err := conn.OpenIOContext(setting.Poolname)
	if err != nil {
		return nil, err
	}

	return &radosdesc{
		Ioctx:     ioctx,
		Conn:      conn,
		Chunksize: chunksize,
	}, nil
}

// file : the path of file to save
func (d *radosdesc) Save(file string) (string, error) {
	fp, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	info, err := fp.Stat()
	if err != nil {
		return "", err
	}
	fileSize := uint64(info.Size())

	buf := make([]byte, d.Chunksize)
	writeUp := false
	totalRead := uint64(0)

	//get object id from Omap
	oid, err := d.getOid(file)
	if err != nil {
		oid = objectBlobPrefix + utils.MD5(uuid.NewV4().String())
		if err = d.setOid(file, oid); err != nil {
			return "", err
		}
	} else {
		// Check total object size only for existing ones
		totalSize, err := d.getXattrTotalSize(oid)
		if err != nil {
			return "", err
		}
		// If new object is smaller, delete old one
		if totalSize > fileSize {
			for offset := uint64(0); offset < totalSize; offset += d.Chunksize {
				chunkName := d.getChunkName(oid, offset)

				err = d.Ioctx.Delete(chunkName)
				if err != nil {
					return "", err
				}
			}
		}
	}

	// Write
	for {
		sizeRead := uint64(0)

		// Read from fp
		for i := 3; sizeRead < d.Chunksize; i-- {
			n, err := fp.Read(buf[sizeRead:])
			sizeRead += uint64(n)
			if err != nil {
				if err != io.EOF {
					return "", err
				}

				writeUp = true
				break
			}

			if i == 0 {
				return "", fmt.Errorf("Not read enough data")
			}
		}

		// End of file and nothing was read
		if sizeRead == 0 {
			break
		}

		// Write chunk object
		chunkName := d.getChunkName(oid, totalRead)
		if err = d.Ioctx.Write(chunkName, buf[:sizeRead], 0); err != nil {
			return "", err
		}
		totalRead += sizeRead

		// Update total object size as xattr in the first chunk of the object
		err = d.setXattrTotalSize(oid, uint64(totalRead))
		if err != nil {
			return "", err
		}

		// End of file
		if writeUp {
			break
		}
	}

	return "", nil
}

func (d *radosdesc) Get(objectPath string) ([]byte, error) {
	// Get oid from filename
	oid, err := d.getOid(objectPath)
	if err != nil {
		return nil, err
	}
	if oid == "" {
		return nil, fmt.Errorf("Is virtual directory not file")
	}

	// Get total size of object from Omap
	totalSize, err := d.getXattrTotalSize(oid)
	if err != nil {
		return nil, err
	}

	buf := make([]byte, totalSize)
	readOffset := uint64(0)

	for readNum := d.Chunksize; readNum == d.Chunksize; {
		// Read chunk object
		chunkName := d.getChunkName(oid, readOffset)
		n, err := d.Ioctx.Read(chunkName, buf[readOffset:], 0)
		if err != nil {
			return nil, err
		}
		readNum = uint64(n)
		readOffset += uint64(readNum)
	}

	if readOffset != totalSize {
		return buf, fmt.Errorf("File corrupted")
	}

	return buf, nil
}

// Delete deletes all objects stored at "path" and its subpaths.
func (d *radosdesc) Delete(objectPath string) error {
	// Get oid
	oid, err := d.getOid(objectPath)
	if err != nil {
		return err
	}

	// Deleting virtual directory
	if oid == "" {
		objects, err := d.Ioctx.GetAllOmapValues(objectPath, "", "", 1)
		if err != nil {
			return err
		}

		for object := range objects {
			err = d.Delete(path.Join(objectPath, object))
			if err != nil {
				return err
			}
		}
	} else {
		// Delete object chunks
		totalSize, err := d.getXattrTotalSize(oid)
		if err != nil {
			return err
		}

		for offset := uint64(0); offset < totalSize; offset += d.Chunksize {
			chunkName := d.getChunkName(oid, offset)

			err = d.Ioctx.Delete(chunkName)
			if err != nil {
				return err
			}
		}

		// Delete reference
		err = d.deleteOid(objectPath)
		if err != nil {
			return err
		}
	}

	return nil
}

// List all objects in driver
func (d *radosdesc) listObjects() ([]string, error) {
	objectList := []string{}
	err := d.Ioctx.ListObjects(func(oid string) {
		objectList = append(objectList, oid)
	})
	return objectList, err
}

// Save object identifier and its hierarchy in Omap
func (d *radosdesc) setOid(objectPath, oid string) error {
	directory := path.Dir(objectPath)
	base := path.Base(objectPath)
	createParentReference := true

	// After creating this reference, skip the parents referencing since the
	// hierarchy already exists
	if oid == "" {
		firstReference, err := d.Ioctx.GetOmapValues(directory, "", "", 1)
		if (err == nil) && (len(firstReference) > 0) {
			createParentReference = false
		}
	}

	oids := map[string][]byte{
		base: []byte(oid),
	}
	// Reference object
	err := d.Ioctx.SetOmap(directory, oids)
	if err != nil {
		return err
	}

	// Esure parent virtual directories
	if createParentReference {
		return d.setOid(directory, "")
	}

	return nil
}

func (d *radosdesc) getOid(objectPath string) (string, error) {
	directory := path.Dir(objectPath)
	base := path.Base(objectPath)

	files, err := d.Ioctx.GetOmapValues(directory, "", base, 1)
	if err != nil {
		return "", err
	}
	if files[base] == nil {
		return "", fmt.Errorf("rados: Not found")
	}

	return string(files[base]), nil
}

// Remove a file from the files hierarchy
func (d *radosdesc) deleteOid(objectPath string) error {
	// Remove object reference
	directory := path.Dir(objectPath)
	base := path.Base(objectPath)
	err := d.Ioctx.RmOmapKeys(directory, []string{base})
	if err != nil {
		return err
	}

	// Remove virtual directory if empty (no more references)
	firstReference, err := d.Ioctx.GetOmapValues(directory, "", "", 1)
	if err != nil {
		return err
	}

	if len(firstReference) == 0 {
		// Delete omap
		err := d.Ioctx.Delete(directory)
		if err != nil {
			return err
		}

		// Remove reference on parent omaps
		if directory != "" {
			return d.deleteOid(directory)
		}
	}

	return nil
}

func (d *radosdesc) getChunkName(oid string, totalRead uint64) string {
	id := totalRead / d.Chunksize
	return oid + "-" + strconv.FormatInt(int64(id), 10)
}

// Set the total size of a chunked object `oid`
func (d *radosdesc) setXattrTotalSize(oid string, size uint64) error {
	// Convert uint64 `size` to []byte
	xattr := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(xattr, size)

	// Save the total size as a xattr in the first chunk
	return d.Ioctx.SetXattr(oid+"-0", defaultXattrTotalSizeName, xattr)
}

func (d *radosdesc) getXattrTotalSize(oid string) (uint64, error) {
	// Fetch xattr as []byte
	xattr := make([]byte, binary.MaxVarintLen64)
	_, err := d.Ioctx.GetXattr(oid+"-0", defaultXattrTotalSizeName, xattr)
	if err != nil {
		return 0, err
	}

	// Convert []byte as uint64
	totalSize := binary.LittleEndian.Uint64(xattr)

	return totalSize, nil
}
