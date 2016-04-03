package rados

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path"
	"strconv"

	"github.com/ceph/go-ceph/rados"

	"github.com/containerops/dockyard/backend/factory"
	"github.com/containerops/dockyard/utils/setting"
)

// Prefix all the stored blob
const objectBlobPrefix = "blob:"

// Stripes objects size to 4M
const defaultChunkSize = 4 << 20
const defaultXattrTotalSizeName = "total-size"

// Max number of keys fetched from omap at each read operation
const defaultKeysFetched = 1

type radosdesc struct{}

func init() {
	factory.Register("rados", &radosdesc{})
}

type radosDriver struct {
	Conn      *rados.Conn
	Ioctx     *rados.IOContext
	Chunksize uint64
}

type driverParameters struct {
	Poolname  string
	Username  string
	Chunksize uint64
}

func new() (*radosDriver, error) {
	chunksize := uint64(defaultChunkSize)
	if setting.Chunksize != "" {
		if tmp, err := strconv.Atoi(setting.Chunksize); err != nil {
			return nil, fmt.Errorf("The chunksize parameter should be a number")
		} else {
			chunksize = uint64(tmp)
		}
	}

	params := driverParameters{
		Poolname:  setting.Poolname,
		Username:  setting.Username,
		Chunksize: chunksize,
	}

	var err error
	var conn *rados.Conn
	if params.Username != "" {
		conn, err = rados.NewConnWithUser(params.Username)
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

	ioctx, err := conn.OpenIOContext(params.Poolname)
	if err != nil {
		return nil, err
	}

	return &radosDriver{
		Ioctx:     ioctx,
		Conn:      conn,
		Chunksize: params.Chunksize,
	}, nil
}

func (r *radosdesc) Save(file string) (string, error) {
	//upload replicated file
	fp, err := os.Open(file)
	if err != nil {
		return "", err
	}
	defer fp.Close()

	d, err := new()
	if err != nil {
		return "", err
	}

	if _, err = d.WriteStream(file, 0, fp); err != nil {
		return "", err
	}

	return "", nil
}

func (r *radosdesc) Put(path string, contents []byte) error {
	d, err := new()
	if err != nil {
		return err
	}

	if _, err = d.WriteStream(path, 0, bytes.NewReader(contents)); err != nil {
		return err
	}

	return nil
}

func (r *radosdesc) Get(path string) ([]byte, error) {
	d, err := new()
	if err != nil {
		return nil, err
	}

	rc, err := d.ReadStream(path, 0)
	if err != nil {
		return nil, err
	}
	defer rc.Close()

	p, err := ioutil.ReadAll(rc)
	if err != nil {
		return nil, err
	}

	return p, nil
}

func (r *radosdesc) Delete(path string) error {
	d, err := new()
	if err != nil {
		return err
	}

	return d.Delete(path)
}

// ReadStream retrieves an io.ReadCloser for the content stored at "path" with a
// given byte offset.
type readStreamReader struct {
	Driver *radosDriver
	Oid    string
	Size   uint64
	Offset uint64
}

func (r *readStreamReader) Close() error {
	return nil
}

func (d *radosDriver) ReadStream(path string, offset int64) (io.ReadCloser, error) {
	// get oid from filename
	oid, err := d.getOid(path)
	if err != nil {
		return nil, err
	}

	// get object stat
	stat, err := d.Stat(path)
	if err != nil {
		return nil, err
	}

	if offset > stat.Size() {
		return nil, InvalidOffsetError{Path: path, Offset: offset}
	}

	return &readStreamReader{
		Driver: d,
		Oid:    oid,
		Size:   uint64(stat.Size()),
		Offset: uint64(offset),
	}, nil
}

func (r *readStreamReader) Read(b []byte) (n int, err error) {
	// Determine the part available to read
	bufferOffset := uint64(0)
	bufferSize := uint64(len(b))

	// End of the object, read less than the buffer size
	if bufferSize > r.Size-r.Offset {
		bufferSize = r.Size - r.Offset
	}

	// Fill `b`
	for bufferOffset < bufferSize {
		// Get the offset in the object chunk
		chunkedOid, chunkedOffset := r.Driver.getChunkNameFromOffset(r.Oid, r.Offset)

		// Determine the best size to read
		bufferEndOffset := bufferSize
		if bufferEndOffset-bufferOffset > r.Driver.Chunksize-chunkedOffset {
			bufferEndOffset = bufferOffset + (r.Driver.Chunksize - chunkedOffset)
		}

		// Read the chunk
		n, err = r.Driver.Ioctx.Read(chunkedOid, b[bufferOffset:bufferEndOffset], chunkedOffset)
		if err != nil {
			return int(bufferOffset), err
		}

		bufferOffset += uint64(n)
		r.Offset += uint64(n)
	}

	// EOF if the offset is at the end of the object
	if r.Offset == r.Size {
		return int(bufferOffset), io.EOF
	}

	return int(bufferOffset), nil
}

func (d *radosDriver) WriteStream(path string, offset int64, reader io.Reader) (totalRead int64, err error) {
	buf := make([]byte, d.Chunksize)
	totalRead = 0

	oid, err := d.getOid(path)
	if err != nil {
		switch err.(type) {
		// Trying to write new object, generate new blob identifier for it
		case PathNotFoundError:
			oid = d.generateOid()
			err = d.putOid(path, oid)
			if err != nil {
				return 0, err
			}
		default:
			return 0, err
		}
	} else {
		// Check total object size only for existing ones
		totalSize, err := d.getXattrTotalSize(oid)
		if err != nil {
			return 0, err
		}

		// If offset if after the current object size, fill the gap with zeros
		for totalSize < uint64(offset) {
			sizeToWrite := d.Chunksize
			if totalSize-uint64(offset) < sizeToWrite {
				sizeToWrite = totalSize - uint64(offset)
			}

			chunkName, chunkOffset := d.getChunkNameFromOffset(oid, uint64(totalSize))
			err = d.Ioctx.Write(chunkName, buf[:sizeToWrite], uint64(chunkOffset))
			if err != nil {
				return totalRead, err
			}

			totalSize += sizeToWrite
		}
	}

	// Writer
	for {
		// Align to chunk size
		sizeRead := uint64(0)
		sizeToRead := uint64(offset+totalRead) % d.Chunksize
		if sizeToRead == 0 {
			sizeToRead = d.Chunksize
		}

		// Read from `reader`
		for sizeRead < sizeToRead {
			nn, err := reader.Read(buf[sizeRead:sizeToRead])
			sizeRead += uint64(nn)
			if err != nil {
				if err != io.EOF {
					return totalRead, err
				}

				break
			}
		}

		// End of file and nothing was read
		if sizeRead == 0 {
			break
		}

		// Write chunk object
		chunkName, chunkOffset := d.getChunkNameFromOffset(oid, uint64(offset+totalRead))
		err = d.Ioctx.Write(chunkName, buf[:sizeRead], uint64(chunkOffset))

		if err != nil {
			return totalRead, err
		}

		// Update total object size as xattr in the first chunk of the object
		err = d.setXattrTotalSize(oid, uint64(offset+totalRead)+sizeRead)
		if err != nil {
			return totalRead, err
		}

		totalRead += int64(sizeRead)

		// End of file
		if sizeRead < sizeToRead {
			break
		}
	}

	return totalRead, nil

}

// Stat retrieves the FileInfo for the given path, including the current size
func (d *radosDriver) Stat(path string) (FileInfo, error) {
	// get oid from filename
	oid, err := d.getOid(path)
	if err != nil {
		return nil, err
	}

	// the path is a virtual directory?
	if oid == "" {
		return FileInfoInternal{
			FileInfoFields: FileInfoFields{
				Path:  path,
				Size:  0,
				IsDir: true,
			},
		}, nil
	}

	// stat first chunk
	stat, err := d.Ioctx.Stat(oid + "-0")
	if err != nil {
		return nil, err
	}

	// get total size of chunked object
	totalSize, err := d.getXattrTotalSize(oid)
	if err != nil {
		return nil, err
	}

	return FileInfoInternal{
		FileInfoFields: FileInfoFields{
			Path:    path,
			Size:    int64(totalSize),
			ModTime: stat.ModTime,
		},
	}, nil
}

// Delete recursively deletes all objects stored at "path" and its subpaths.
func (d *radosDriver) Delete(objectPath string) error {
	// Get oid
	oid, err := d.getOid(objectPath)
	if err != nil {
		return err
	}

	// Deleting virtual directory
	if oid == "" {
		objects, err := d.listDirectoryOid(objectPath)
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
			chunkName, _ := d.getChunkNameFromOffset(oid, offset)

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

// Generate a blob identifier
func (d *radosDriver) generateOid() string {
	return objectBlobPrefix + Generate().String()
}

// Reference a object and its hierarchy
func (d *radosDriver) putOid(objectPath string, oid string) error {
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
		return d.putOid(directory, "")
	}

	return nil
}

func (d *radosDriver) getOid(objectPath string) (string, error) {
	directory := path.Dir(objectPath)
	base := path.Base(objectPath)

	files, err := d.Ioctx.GetOmapValues(directory, "", base, 1)
	if (err != nil) || (files[base] == nil) {
		return "", PathNotFoundError{Path: objectPath}
	}

	return string(files[base]), nil
}

// List the objects of a virtual directory
func (d *radosDriver) listDirectoryOid(path string) (list map[string][]byte, err error) {
	return d.Ioctx.GetAllOmapValues(path, "", "", defaultKeysFetched)
}

// Remove a file from the files hierarchy
func (d *radosDriver) deleteOid(objectPath string) error {
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

// Takes an offset in an chunked object and return the chunk name and a new
// offset in this chunk object
func (d *radosDriver) getChunkNameFromOffset(oid string, offset uint64) (string, uint64) {
	chunkID := offset / d.Chunksize
	chunkedOid := oid + "-" + strconv.FormatInt(int64(chunkID), 10)
	chunkedOffset := offset % d.Chunksize

	return chunkedOid, chunkedOffset
}

// Set the total size of a chunked object `oid`
func (d *radosDriver) setXattrTotalSize(oid string, size uint64) error {
	// Convert uint64 `size` to []byte
	xattr := make([]byte, binary.MaxVarintLen64)
	binary.LittleEndian.PutUint64(xattr, size)

	// Save the total size as a xattr in the first chunk
	return d.Ioctx.SetXattr(oid+"-0", defaultXattrTotalSizeName, xattr)
}

func (d *radosDriver) getXattrTotalSize(oid string) (uint64, error) {
	// Fetch xattr as []byte
	xattr := make([]byte, binary.MaxVarintLen64)
	xattrLength, err := d.Ioctx.GetXattr(oid+"-0", defaultXattrTotalSizeName, xattr)
	if err != nil {
		return 0, err
	}

	if xattrLength != len(xattr) {
		fmt.Printf("object %s xattr length mismatch: %d != %d", oid, xattrLength, len(xattr))
		return 0, PathNotFoundError{Path: oid}
	}

	// Convert []byte as uint64
	totalSize := binary.LittleEndian.Uint64(xattr)

	return totalSize, nil
}
