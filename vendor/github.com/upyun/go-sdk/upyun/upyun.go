// package upyun is used for your UPYUN bucket
// this sdk implement purge api, form api, http rest api
package upyun

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"fmt"
	"io"
	"net"
	"net/http"
	URL "net/url"
	"strconv"
	"strings"
	"time"
)

const (
	Version = "2.0.0"
)

const (
	// Default(Min/Max)ChunkSize: set the buffer size when doing copy operation
	defaultChunkSize = 32 * 1024
	// defaultConnectTimeout: connection timeout when connect to upyun endpoint
	defaultConnectTimeout = 60
)

// chunkSize: chunk size when copy
var (
	chunkSize = defaultChunkSize
)

// Util functions

// User Agent
func makeUserAgent() string {
	return fmt.Sprintf("UPYUN Go SDK %s", Version)
}

// Greenwich Mean Time
func genRFC1123Date() string {
	return time.Now().UTC().Format(time.RFC1123)
}

// make md5 from string
func md5Str(s string) (ret string) {
	return fmt.Sprintf("%x", md5.Sum([]byte(s)))
}

// make base64 from []byte
func base64Str(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// URL encode
func encodeURL(uri string) string {
	return base64.URLEncoding.EncodeToString([]byte(uri))
}

// URI escape
func escapeURI(uri string) string {
	Uri := URL.URL{}
	Uri.Path = uri
	return Uri.String()
}

func md5sum(fd io.Reader) (string, int64, error) {
	var result []byte
	hash := md5.New()
	if written, err := io.Copy(hash, fd); err != nil {
		return "", written, err
	} else {
		return hex.EncodeToString(hash.Sum(result)), written, nil
	}
}

// Because of io.Copy use a 32Kb buffer, and, it is hard coded
// user can specify a chunksize with upyun.SetChunkSize
func chunkedCopy(dst io.Writer, src io.Reader) (written int64, err error) {
	buf := make([]byte, chunkSize)

	for {
		nr, er := src.Read(buf)
		if nr > 0 {
			nw, ew := dst.Write(buf[0:nr])

			if nw > 0 {
				written += int64(nw)
			}
			if ew != nil {
				err = ew
				break
			}
			if nr != nw {
				err = io.ErrShortWrite
				break
			}
		}
		if er == io.EOF {
			break
		}
		if er != nil {
			err = er
			break
		}
	}
	return
}

// Use for http connection timeout
func timeoutDialer(timeout int) func(string, string) (net.Conn, error) {
	return func(network, addr string) (c net.Conn, err error) {
		c, err = net.DialTimeout(network, addr, time.Duration(timeout)*time.Second)
		if err != nil {
			return nil, err
		}
		return
	}
}

func SetChunkSize(chunksize int) {
	chunkSize = chunksize
}

// FileInfo when use getlist
type FileInfo struct {
	Size int64
	Time time.Time
	Name string
	Type string
}

func newFileInfo(arg interface{}) *FileInfo {
	switch arg.(type) {
	case string:
		s := arg.(string)
		infoList := strings.Split(s, "\t")
		if len(infoList) != 4 {
			return nil
		}

		size, _ := strconv.ParseInt(infoList[2], 10, 64)
		timestamp, _ := strconv.ParseInt(infoList[3], 10, 64)
		typ := "folder"
		if infoList[1] != "F" {
			typ = "file"
		}

		return &FileInfo{
			Name: infoList[0],
			Type: typ,
			Size: size,
			Time: time.Unix(timestamp, 0),
		}

	default:
		var fileInfo FileInfo
		headers := arg.(http.Header)
		for k, v := range headers {
			switch {
			case strings.Contains(k, "File-Type"):
				fileInfo.Type = v[0]
			case strings.Contains(k, "File-Size"):
				fileInfo.Size, _ = strconv.ParseInt(v[0], 10, 64)
			case strings.Contains(k, "File-Date"):
				timestamp, _ := strconv.ParseInt(v[0], 10, 64)
				fileInfo.Time = time.Unix(timestamp, 0)
			}
		}
		return &fileInfo
	}
}
