package utils

import (
	"bufio"
	"bytes"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/docker/docker/pkg/archive"
	"github.com/docker/docker/pkg/fileutils"
	"github.com/docker/docker/pkg/stdcopy"
)

//****************************************************************//
//auth use type
//****************************************************************//

// AuthConfiguration represents authentication options to use in the PushImage
// method. It represents the authentication in the Docker index server.
type AuthConfiguration struct {
	Username      string `json:"username,omitempty"`
	Password      string `json:"password,omitempty"`
	Email         string `json:"email,omitempty"`
	ServerAddress string `json:"serveraddress,omitempty"`
}

// AuthConfigurations represents authentication options to use for the
// PushImage method accommodating the new X-Registry-Config header
type AuthConfigurations struct {
	Configs map[string]AuthConfiguration `json:"configs"`
}

// dockerConfig represents a registry authentation configuration from the
// .dockercfg file.
type dockerConfig struct {
	Auth  string `json:"auth"`
	Email string `json:"email"`
}

//****************************************************************//
//change use type
//****************************************************************//

// ChangeType is a type for constants indicating the type of change
// in a container
type ChangeType int

// Change represents a change in a container.
//
// See  https://docs.docker.com/reference/api/docker_remote_api_v1.19/#inspect-changes-on-a-container-s-filesystem for more details.
type Change struct {
	Path string
	Kind ChangeType
}

//****************************************************************//
//clent use type
//****************************************************************//
// APIVersion is an internal representation of a version of the Remote API.
type APIVersion []int

// Client is the basic type of this package. It provides methods for
// interaction with the API.
type Client struct {
	SkipServerVersionCheck bool
	HTTPClient             *http.Client
	transport              *http.Transport
	TLSConfig              *tls.Config

	endpoint            string
	endpointURL         *url.URL
	eventMonitor        *eventMonitoringState
	requestedAPIVersion APIVersion
	serverAPIVersion    APIVersion
	expectedAPIVersion  APIVersion
}

type doOptions struct {
	data      interface{}
	forceJSON bool
}

type streamOptions struct {
	setRawTerminal bool
	rawJSONStream  bool
	useJSONDecoder bool
	headers        map[string]string
	in             io.Reader
	stdout         io.Writer
	stderr         io.Writer
}

type hijackOptions struct {
	success        chan struct{}
	setRawTerminal bool
	in             io.Reader
	stdout         io.Writer
	stderr         io.Writer
	data           interface{}
}

type jsonMessage struct {
	Status   string `json:"status,omitempty"`
	Progress string `json:"progress,omitempty"`
	Error    string `json:"error,omitempty"`
	Stream   string `json:"stream,omitempty"`
}

// Error represents failures in the API. It represents a failure from the API.
type Error struct {
	Status  int
	Message string
}

//****************************************************************//
//Container use type
//****************************************************************//

// ListContainersOptions specify parameters to the ListContainers function.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#list-containers for more details.
/*
List containers
GET /containers/json
Example request:
	GET /containers/json?all=1&before=8dfafdbc3a40&size=1 HTTP/1.1
Example response:
	HTTP/1.1 200 OK
	Content-Type: application/json
[
     {
             "Id": "8dfafdbc3a40",
             "Image": "ubuntu:latest",
             "Command": "echo 1",
             "Created": 1367854155,
             "Status": "Exit 0",
             "Ports": [{"PrivatePort": 2222, "PublicPort": 3333, "Type": "tcp"}],
             "SizeRw": 12288,
             "SizeRootFs": 0
     },
	 ......
]
Query Parameters:
	all – 1/True/true or 0/False/false, Show all containers. Only running containers are shown by default (i.e., this defaults to false)
	limit – Show limit last created containers, include non-running ones.
	since – Show only containers created since Id, include non-running ones.
	before – Show only containers created before Id, include non-running ones.
	size – 1/True/true or 0/False/false, Show the containers sizes
	filters - a JSON encoded value of the filters (a map[string][]string) to process on the containers list. Available filters:
	exited=<int>; – containers with exit code of <int> ;
	status=(restarting|running|paused|exited)
	label=key or key=value of a container label
Status Codes:
	200 – no error
	400 – bad parameter
	500 – server error
*/
type ListContainersOptions struct {
	All     bool
	Limit   int
	Since   string
	Before  string
	Size    bool
	Filters map[string][]string
	Exited  int
	Status  string
	Label   string
	Key     string
}

// APIPort is a type that represents a port mapping returned by the Docker API
type APIPort struct {
	PrivatePort int64  `json:"PrivatePort,omitempty" yaml:"PrivatePort,omitempty"`
	PublicPort  int64  `json:"PublicPort,omitempty" yaml:"PublicPort,omitempty"`
	Type        string `json:"Type,omitempty" yaml:"Type,omitempty"`
	IP          string `json:"IP,omitempty" yaml:"IP,omitempty"`
}

// APIContainers represents a container.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#list-containers for more details.
type APIContainers struct {
	ID         string    `json:"Id" yaml:"Id"`
	Image      string    `json:"Image,omitempty" yaml:"Image,omitempty"`
	Command    string    `json:"Command,omitempty" yaml:"Command,omitempty"`
	Created    int64     `json:"Created,omitempty" yaml:"Created,omitempty"`
	Status     string    `json:"Status,omitempty" yaml:"Status,omitempty"`
	Ports      []APIPort `json:"Ports,omitempty" yaml:"Ports,omitempty"`
	SizeRw     int64     `json:"SizeRw,omitempty" yaml:"SizeRw,omitempty"`
	SizeRootFs int64     `json:"SizeRootFs,omitempty" yaml:"SizeRootFs,omitempty"`
	Names      []string  `json:"Names,omitempty" yaml:"Names,omitempty"` //is exist???
}

// Port represents the port number and the protocol, in the form
// <number>/<protocol>. For example: 80/tcp.
type Port string

// State represents the state of a container.
type State struct {
	Running    bool      `json:"Running,omitempty" yaml:"Running,omitempty"`
	Paused     bool      `json:"Paused,omitempty" yaml:"Paused,omitempty"`
	Restarting bool      `json:"Restarting,omitempty" yaml:"Restarting,omitempty"`
	OOMKilled  bool      `json:"OOMKilled,omitempty" yaml:"OOMKilled,omitempty"`
	Pid        int       `json:"Pid,omitempty" yaml:"Pid,omitempty"`
	ExitCode   int       `json:"ExitCode,omitempty" yaml:"ExitCode,omitempty"`
	Error      string    `json:"Error,omitempty" yaml:"Error,omitempty"`
	StartedAt  time.Time `json:"StartedAt,omitempty" yaml:"StartedAt,omitempty"`
	FinishedAt time.Time `json:"FinishedAt,omitempty" yaml:"FinishedAt,omitempty"`
}

// PortBinding represents the host/container port mapping as returned in the
// `docker inspect` json
type PortBinding struct {
	HostIP   string `json:"HostIP,omitempty" yaml:"HostIP,omitempty"`
	HostPort string `json:"HostPort,omitempty" yaml:"HostPort,omitempty"`
}

// PortMapping represents a deprecated field in the `docker inspect` output,
// and its value as found in NetworkSettings should always be nil
type PortMapping map[string]string

// NetworkSettings contains network-related information about a container
type NetworkSettings struct {
	IPAddress   string                 `json:"IPAddress,omitempty" yaml:"IPAddress,omitempty"`
	IPPrefixLen int                    `json:"IPPrefixLen,omitempty" yaml:"IPPrefixLen,omitempty"`
	Gateway     string                 `json:"Gateway,omitempty" yaml:"Gateway,omitempty"`
	Bridge      string                 `json:"Bridge,omitempty" yaml:"Bridge,omitempty"`
	PortMapping map[string]PortMapping `json:"PortMapping,omitempty" yaml:"PortMapping,omitempty"`
	Ports       map[Port][]PortBinding `json:"Ports,omitempty" yaml:"Ports,omitempty"`
}

// Config is the list of configuration options used when creating a container.
// Config does not contain the options that are specific to starting a container on a
// given host.  Those are contained in HostConfig
type Config struct {
	Hostname        string              `json:"Hostname,omitempty" yaml:"Hostname,omitempty"`
	Domainname      string              `json:"Domainname,omitempty" yaml:"Domainname,omitempty"`
	User            string              `json:"User,omitempty" yaml:"User,omitempty"`
	Memory          int64               `json:"Memory,omitempty" yaml:"Memory,omitempty"`
	MemorySwap      int64               `json:"MemorySwap,omitempty" yaml:"MemorySwap,omitempty"`
	CPUShares       int64               `json:"CpuShares,omitempty" yaml:"CpuShares,omitempty"`
	CPUSet          string              `json:"Cpuset,omitempty" yaml:"Cpuset,omitempty"`
	AttachStdin     bool                `json:"AttachStdin,omitempty" yaml:"AttachStdin,omitempty"`
	AttachStdout    bool                `json:"AttachStdout,omitempty" yaml:"AttachStdout,omitempty"`
	AttachStderr    bool                `json:"AttachStderr,omitempty" yaml:"AttachStderr,omitempty"`
	PortSpecs       []string            `json:"PortSpecs,omitempty" yaml:"PortSpecs,omitempty"`
	ExposedPorts    map[Port]struct{}   `json:"ExposedPorts,omitempty" yaml:"ExposedPorts,omitempty"`
	Tty             bool                `json:"Tty,omitempty" yaml:"Tty,omitempty"`
	OpenStdin       bool                `json:"OpenStdin,omitempty" yaml:"OpenStdin,omitempty"`
	StdinOnce       bool                `json:"StdinOnce,omitempty" yaml:"StdinOnce,omitempty"`
	Env             []string            `json:"Env,omitempty" yaml:"Env,omitempty"`
	Cmd             []string            `json:"Cmd,omitempty" yaml:"Cmd,omitempty"`
	DNS             []string            `json:"Dns,omitempty" yaml:"Dns,omitempty"` // For Docker API v1.9 and below only
	Image           string              `json:"Image,omitempty" yaml:"Image,omitempty"`
	Volumes         map[string]struct{} `json:"Volumes,omitempty" yaml:"Volumes,omitempty"`
	VolumesFrom     string              `json:"VolumesFrom,omitempty" yaml:"VolumesFrom,omitempty"`
	WorkingDir      string              `json:"WorkingDir,omitempty" yaml:"WorkingDir,omitempty"`
	MacAddress      string              `json:"MacAddress,omitempty" yaml:"MacAddress,omitempty"`
	Entrypoint      []string            `json:"Entrypoint,omitempty" yaml:"Entrypoint,omitempty"`
	NetworkDisabled bool                `json:"NetworkDisabled,omitempty" yaml:"NetworkDisabled,omitempty"`
	SecurityOpts    []string            `json:"SecurityOpts,omitempty" yaml:"SecurityOpts,omitempty"`
	OnBuild         []string            `json:"OnBuild,omitempty" yaml:"OnBuild,omitempty"`
	Labels          map[string]string   `json:"Labels,omitempty" yaml:"Labels,omitempty"`
}

// LogConfig defines the log driver type and the configuration for it.
type LogConfig struct {
	Type   string            `json:"Type,omitempty" yaml:"Type,omitempty"`
	Config map[string]string `json:"Config,omitempty" yaml:"Config,omitempty"`
}

// ULimit defines system-wide resource limitations
// This can help a lot in system administration, e.g. when a user starts too many processes and therefore makes the system unresponsive for other users.
type ULimit struct {
	Name string `json:"Name,omitempty" yaml:"Name,omitempty"`
	Soft int64  `json:"Soft,omitempty" yaml:"Soft,omitempty"`
	Hard int64  `json:"Hard,omitempty" yaml:"Hard,omitempty"`
}

// SwarmNode containers information about which Swarm node the container is on
type SwarmNode struct {
	ID     string            `json:"ID,omitempty" yaml:"ID,omitempty"`
	IP     string            `json:"IP,omitempty" yaml:"IP,omitempty"`
	Addr   string            `json:"Addr,omitempty" yaml:"Addr,omitempty"`
	Name   string            `json:"Name,omitempty" yaml:"Name,omitempty"`
	CPUs   int64             `json:"CPUs,omitempty" yaml:"CPUs,omitempty"`
	Memory int64             `json:"Memory,omitempty" yaml:"Memory,omitempty"`
	Labels map[string]string `json:"Labels,omitempty" yaml:"Labels,omitempty"`
}

// Container is the type encompasing everything about a container - its config,
// hostconfig, etc.
type Container struct {
	ID string `json:"Id" yaml:"Id"`

	Created time.Time `json:"Created,omitempty" yaml:"Created,omitempty"`

	Path string   `json:"Path,omitempty" yaml:"Path,omitempty"`
	Args []string `json:"Args,omitempty" yaml:"Args,omitempty"`

	Config *Config `json:"Config,omitempty" yaml:"Config,omitempty"`
	State  State   `json:"State,omitempty" yaml:"State,omitempty"`
	Image  string  `json:"Image,omitempty" yaml:"Image,omitempty"`

	Node *SwarmNode `json:"Node,omitempty" yaml:"Node,omitempty"`

	NetworkSettings *NetworkSettings `json:"NetworkSettings,omitempty" yaml:"NetworkSettings,omitempty"`

	SysInitPath    string `json:"SysInitPath,omitempty" yaml:"SysInitPath,omitempty"`
	ResolvConfPath string `json:"ResolvConfPath,omitempty" yaml:"ResolvConfPath,omitempty"`
	HostnamePath   string `json:"HostnamePath,omitempty" yaml:"HostnamePath,omitempty"`
	HostsPath      string `json:"HostsPath,omitempty" yaml:"HostsPath,omitempty"`
	Name           string `json:"Name,omitempty" yaml:"Name,omitempty"`
	Driver         string `json:"Driver,omitempty" yaml:"Driver,omitempty"`

	Volumes    map[string]string `json:"Volumes,omitempty" yaml:"Volumes,omitempty"`
	VolumesRW  map[string]bool   `json:"VolumesRW,omitempty" yaml:"VolumesRW,omitempty"`
	HostConfig *HostConfig       `json:"HostConfig,omitempty" yaml:"HostConfig,omitempty"`
	ExecIDs    []string          `json:"ExecIDs,omitempty" yaml:"ExecIDs,omitempty"`

	RestartCount int `json:"RestartCount,omitempty" yaml:"RestartCount,omitempty"`

	AppArmorProfile string `json:"AppArmorProfile,omitempty" yaml:"AppArmorProfile,omitempty"`
}

// RenameContainerOptions specify parameters to the RenameContainer function.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#rename-a-container for more details.
/*
-------------------------------------------------------------------------------------
Rename a container
POST /containers/(id)/rename
Rename the container id to a new_name
Example request:
	POST /containers/e90e34656806/rename?name=new_name HTTP/1.1
Example response:
	HTTP/1.1 204 No Content

Query Parameters:
	name – new name for the container
	Status Codes:
		204 – no error
		404 – no such container
		409 - conflict name already assigned
		500 – server error
-------------------------------------------------------------------------------------
*/

type RenameContainerOptions struct {
	// ID of container to rename
	ID string `qs:"-"`

	// New name
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
}

// CreateContainerOptions specify parameters to the CreateContainer function.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#create-a-container for more details.
type CreateContainerOptions struct {
	Name       string
	Config     *Config `qs:"-"`
	HostConfig *HostConfig
}

// KeyValuePair is a type for generic key/value pairs as used in the Lxc
// configuration
type KeyValuePair struct {
	Key   string `json:"Key,omitempty" yaml:"Key,omitempty"`
	Value string `json:"Value,omitempty" yaml:"Value,omitempty"`
}

// RestartPolicy represents the policy for automatically restarting a container.
//
// Possible values are:
//
//   - always: the docker daemon will always restart the container
//   - on-failure: the docker daemon will restart the container on failures, at
//                 most MaximumRetryCount times
//   - no: the docker daemon will not restart the container automatically
type RestartPolicy struct {
	Name              string `json:"Name,omitempty" yaml:"Name,omitempty"`
	MaximumRetryCount int    `json:"MaximumRetryCount,omitempty" yaml:"MaximumRetryCount,omitempty"`
}

// Device represents a device mapping between the Docker host and the
// container.
type Device struct {
	PathOnHost        string `json:"PathOnHost,omitempty" yaml:"PathOnHost,omitempty"`
	PathInContainer   string `json:"PathInContainer,omitempty" yaml:"PathInContainer,omitempty"`
	CgroupPermissions string `json:"CgroupPermissions,omitempty" yaml:"CgroupPermissions,omitempty"`
}

// HostConfig contains the container options related to starting a container on
// a given host
type HostConfig struct {
	Binds           []string               `json:"Binds,omitempty" yaml:"Binds,omitempty"`
	CapAdd          []string               `json:"CapAdd,omitempty" yaml:"CapAdd,omitempty"`
	CapDrop         []string               `json:"CapDrop,omitempty" yaml:"CapDrop,omitempty"`
	ContainerIDFile string                 `json:"ContainerIDFile,omitempty" yaml:"ContainerIDFile,omitempty"`
	LxcConf         []KeyValuePair         `json:"LxcConf,omitempty" yaml:"LxcConf,omitempty"`
	Privileged      bool                   `json:"Privileged,omitempty" yaml:"Privileged,omitempty"`
	PortBindings    map[Port][]PortBinding `json:"PortBindings,omitempty" yaml:"PortBindings,omitempty"`
	Links           []string               `json:"Links,omitempty" yaml:"Links,omitempty"`
	PublishAllPorts bool                   `json:"PublishAllPorts,omitempty" yaml:"PublishAllPorts,omitempty"`
	DNS             []string               `json:"Dns,omitempty" yaml:"Dns,omitempty"` // For Docker API v1.10 and above only
	DNSSearch       []string               `json:"DnsSearch,omitempty" yaml:"DnsSearch,omitempty"`
	ExtraHosts      []string               `json:"ExtraHosts,omitempty" yaml:"ExtraHosts,omitempty"`
	VolumesFrom     []string               `json:"VolumesFrom,omitempty" yaml:"VolumesFrom,omitempty"`
	NetworkMode     string                 `json:"NetworkMode,omitempty" yaml:"NetworkMode,omitempty"`
	IpcMode         string                 `json:"IpcMode,omitempty" yaml:"IpcMode,omitempty"`
	PidMode         string                 `json:"PidMode,omitempty" yaml:"PidMode,omitempty"`
	RestartPolicy   RestartPolicy          `json:"RestartPolicy,omitempty" yaml:"RestartPolicy,omitempty"`
	Devices         []Device               `json:"Devices,omitempty" yaml:"Devices,omitempty"`
	LogConfig       LogConfig              `json:"LogConfig,omitempty" yaml:"LogConfig,omitempty"`
	ReadonlyRootfs  bool                   `json:"ReadonlyRootfs,omitempty" yaml:"ReadonlyRootfs,omitempty"`
	SecurityOpt     []string               `json:"SecurityOpt,omitempty" yaml:"SecurityOpt,omitempty"`
	CgroupParent    string                 `json:"CgroupParent,omitempty" yaml:"CgroupParent,omitempty"`
	Memory          int64                  `json:"Memory,omitempty" yaml:"Memory,omitempty"`
	MemorySwap      int64                  `json:"MemorySwap,omitempty" yaml:"MemorySwap,omitempty"`
	CPUShares       int64                  `json:"CpuShares,omitempty" yaml:"CpuShares,omitempty"`
	CPUSet          string                 `json:"Cpuset,omitempty" yaml:"Cpuset,omitempty"`
	CPUQuota        int64                  `json:"CpuQuota,omitempty" yaml:"CpuQuota,omitempty"`
	CPUPeriod       int64                  `json:"CpuPeriod,omitempty" yaml:"CpuPeriod,omitempty"`
	Ulimits         []ULimit               `json:"Ulimits,omitempty" yaml:"Ulimits,omitempty"`
}

// TopResult represents the list of processes running in a container, as
// returned by /containers/<id>/top.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#list-processes-running-inside-a-container for more details.
type TopResult struct {
	Titles    []string
	Processes [][]string
}

// Stats represents container statistics, returned by /containers/<id>/stats.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#get-container-stats-based-on-resource-usage for more details.
type Stats struct {
	Read    time.Time `json:"read,omitempty" yaml:"read,omitempty"`
	Network struct {
		RxDropped uint64 `json:"rx_dropped,omitempty" yaml:"rx_dropped,omitempty"`
		RxBytes   uint64 `json:"rx_bytes,omitempty" yaml:"rx_bytes,omitempty"`
		RxErrors  uint64 `json:"rx_errors,omitempty" yaml:"rx_errors,omitempty"`
		TxPackets uint64 `json:"tx_packets,omitempty" yaml:"tx_packets,omitempty"`
		TxDropped uint64 `json:"tx_dropped,omitempty" yaml:"tx_dropped,omitempty"`
		RxPackets uint64 `json:"rx_packets,omitempty" yaml:"rx_packets,omitempty"`
		TxErrors  uint64 `json:"tx_errors,omitempty" yaml:"tx_errors,omitempty"`
		TxBytes   uint64 `json:"tx_bytes,omitempty" yaml:"tx_bytes,omitempty"`
	} `json:"network,omitempty" yaml:"network,omitempty"`
	MemoryStats struct {
		Stats struct {
			TotalPgmafault          uint64 `json:"total_pgmafault,omitempty" yaml:"total_pgmafault,omitempty"`
			Cache                   uint64 `json:"cache,omitempty" yaml:"cache,omitempty"`
			MappedFile              uint64 `json:"mapped_file,omitempty" yaml:"mapped_file,omitempty"`
			TotalInactiveFile       uint64 `json:"total_inactive_file,omitempty" yaml:"total_inactive_file,omitempty"`
			Pgpgout                 uint64 `json:"pgpgout,omitempty" yaml:"pgpgout,omitempty"`
			Rss                     uint64 `json:"rss,omitempty" yaml:"rss,omitempty"`
			TotalMappedFile         uint64 `json:"total_mapped_file,omitempty" yaml:"total_mapped_file,omitempty"`
			Writeback               uint64 `json:"writeback,omitempty" yaml:"writeback,omitempty"`
			Unevictable             uint64 `json:"unevictable,omitempty" yaml:"unevictable,omitempty"`
			Pgpgin                  uint64 `json:"pgpgin,omitempty" yaml:"pgpgin,omitempty"`
			TotalUnevictable        uint64 `json:"total_unevictable,omitempty" yaml:"total_unevictable,omitempty"`
			Pgmajfault              uint64 `json:"pgmajfault,omitempty" yaml:"pgmajfault,omitempty"`
			TotalRss                uint64 `json:"total_rss,omitempty" yaml:"total_rss,omitempty"`
			TotalRssHuge            uint64 `json:"total_rss_huge,omitempty" yaml:"total_rss_huge,omitempty"`
			TotalWriteback          uint64 `json:"total_writeback,omitempty" yaml:"total_writeback,omitempty"`
			TotalInactiveAnon       uint64 `json:"total_inactive_anon,omitempty" yaml:"total_inactive_anon,omitempty"`
			RssHuge                 uint64 `json:"rss_huge,omitempty" yaml:"rss_huge,omitempty"`
			HierarchicalMemoryLimit uint64 `json:"hierarchical_memory_limit,omitempty" yaml:"hierarchical_memory_limit,omitempty"`
			TotalPgfault            uint64 `json:"total_pgfault,omitempty" yaml:"total_pgfault,omitempty"`
			TotalActiveFile         uint64 `json:"total_active_file,omitempty" yaml:"total_active_file,omitempty"`
			ActiveAnon              uint64 `json:"active_anon,omitempty" yaml:"active_anon,omitempty"`
			TotalActiveAnon         uint64 `json:"total_active_anon,omitempty" yaml:"total_active_anon,omitempty"`
			TotalPgpgout            uint64 `json:"total_pgpgout,omitempty" yaml:"total_pgpgout,omitempty"`
			TotalCache              uint64 `json:"total_cache,omitempty" yaml:"total_cache,omitempty"`
			InactiveAnon            uint64 `json:"inactive_anon,omitempty" yaml:"inactive_anon,omitempty"`
			ActiveFile              uint64 `json:"active_file,omitempty" yaml:"active_file,omitempty"`
			Pgfault                 uint64 `json:"pgfault,omitempty" yaml:"pgfault,omitempty"`
			InactiveFile            uint64 `json:"inactive_file,omitempty" yaml:"inactive_file,omitempty"`
			TotalPgpgin             uint64 `json:"total_pgpgin,omitempty" yaml:"total_pgpgin,omitempty"`
		} `json:"stats,omitempty" yaml:"stats,omitempty"`
		MaxUsage uint64 `json:"max_usage,omitempty" yaml:"max_usage,omitempty"`
		Usage    uint64 `json:"usage,omitempty" yaml:"usage,omitempty"`
		Failcnt  uint64 `json:"failcnt,omitempty" yaml:"failcnt,omitempty"`
		Limit    uint64 `json:"limit,omitempty" yaml:"limit,omitempty"`
	} `json:"memory_stats,omitempty" yaml:"memory_stats,omitempty"`
	BlkioStats struct {
		IOServiceBytesRecursive []BlkioStatsEntry `json:"io_service_bytes_recursive,omitempty" yaml:"io_service_bytes_recursive,omitempty"`
		IOServicedRecursive     []BlkioStatsEntry `json:"io_serviced_recursive,omitempty" yaml:"io_serviced_recursive,omitempty"`
		IOQueueRecursive        []BlkioStatsEntry `json:"io_queue_recursive,omitempty" yaml:"io_queue_recursive,omitempty"`
		IOServiceTimeRecursive  []BlkioStatsEntry `json:"io_service_time_recursive,omitempty" yaml:"io_service_time_recursive,omitempty"`
		IOWaitTimeRecursive     []BlkioStatsEntry `json:"io_wait_time_recursive,omitempty" yaml:"io_wait_time_recursive,omitempty"`
		IOMergedRecursive       []BlkioStatsEntry `json:"io_merged_recursive,omitempty" yaml:"io_merged_recursive,omitempty"`
		IOTimeRecursive         []BlkioStatsEntry `json:"io_time_recursive,omitempty" yaml:"io_time_recursive,omitempty"`
		SectorsRecursive        []BlkioStatsEntry `json:"sectors_recursive,omitempty" yaml:"sectors_recursive,omitempty"`
	} `json:"blkio_stats,omitempty" yaml:"blkio_stats,omitempty"`
	CPUStats struct {
		CPUUsage struct {
			PercpuUsage       []uint64 `json:"percpu_usage,omitempty" yaml:"percpu_usage,omitempty"`
			UsageInUsermode   uint64   `json:"usage_in_usermode,omitempty" yaml:"usage_in_usermode,omitempty"`
			TotalUsage        uint64   `json:"total_usage,omitempty" yaml:"total_usage,omitempty"`
			UsageInKernelmode uint64   `json:"usage_in_kernelmode,omitempty" yaml:"usage_in_kernelmode,omitempty"`
		} `json:"cpu_usage,omitempty" yaml:"cpu_usage,omitempty"`
		SystemCPUUsage uint64 `json:"system_cpu_usage,omitempty" yaml:"system_cpu_usage,omitempty"`
		ThrottlingData struct {
			Periods          uint64 `json:"periods,omitempty"`
			ThrottledPeriods uint64 `json:"throttled_periods,omitempty"`
			ThrottledTime    uint64 `json:"throttled_time,omitempty"`
		} `json:"throttling_data,omitempty" yaml:"throttling_data,omitempty"`
	} `json:"cpu_stats,omitempty" yaml:"cpu_stats,omitempty"`
}

// BlkioStatsEntry is a stats entry for blkio_stats
type BlkioStatsEntry struct {
	Major uint64 `json:"major,omitempty" yaml:"major,omitempty"`
	Minor uint64 `json:"minor,omitempty" yaml:"minor,omitempty"`
	Op    string `json:"op,omitempty" yaml:"op,omitempty"`
	Value uint64 `json:"value,omitempty" yaml:"value,omitempty"`
}

// StatsOptions specify parameters to the Stats function.
//
// See http://goo.gl/DFMiYD for more details.
type StatsOptions struct {
	ID     string
	Stats  chan<- *Stats
	Stream bool
}

// KillContainerOptions represents the set of options that can be used in a
// call to KillContainer.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#kill-a-container for more details.
type KillContainerOptions struct {
	// The ID of the container.
	ID string `qs:"-"`

	// The signal to send to the container. When omitted, Docker server
	// will assume SIGKILL.
	Signal Signal
}

// RemoveContainerOptions encapsulates options to remove a container.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#remove-a-container for more details.
type RemoveContainerOptions struct {
	// The ID of the container.
	ID string `qs:"-"`

	// A flag that indicates whether Docker should remove the volumes
	// associated to the container.
	RemoveVolumes bool `qs:"v"`

	// A flag that indicates whether Docker should remove the container
	// even if it is currently running.
	Force bool
}

// CopyFromContainerOptions is the set of options that can be used when copying
// files or folders from a container.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#copy-files-or-folders-from-a-container for more details.
type CopyFromContainerOptions struct {
	OutputStream io.Writer `json:"-"`
	Container    string    `json:"-"`
	Resource     string
}

// CommitContainerOptions aggregates parameters to the CommitContainer method.
//
// See http://goo.gl/Jn8pe8 for more details.
type CommitContainerOptions struct {
	Container  string
	Repository string `qs:"repo"`
	Tag        string
	Message    string `qs:"m"`
	Author     string
	Run        *Config `qs:"-"`
}

// AttachToContainerOptions is the set of options that can be used when
// attaching to a container.
//
// See http://goo.gl/RRAhws for more details.
type AttachToContainerOptions struct {
	Container    string    `qs:"-"`
	InputStream  io.Reader `qs:"-"`
	OutputStream io.Writer `qs:"-"`
	ErrorStream  io.Writer `qs:"-"`

	// Get container logs, sending it to OutputStream.
	Logs bool

	// Stream the response?
	Stream bool

	// Attach to stdin, and use InputStream.
	Stdin bool

	// Attach to stdout, and use OutputStream.
	Stdout bool

	// Attach to stderr, and use ErrorStream.
	Stderr bool

	// If set, after a successful connect, a sentinel will be sent and then the
	// client will block on receive before continuing.
	//
	// It must be an unbuffered channel. Using a buffered channel can lead
	// to unexpected behavior.
	Success chan struct{}

	// Use raw terminal? Usually true when the container contains a TTY.
	RawTerminal bool `qs:"-"`
}

// LogsOptions represents the set of options used when getting logs from a
// container.
//
// See http://goo.gl/rLhKSU for more details.
type LogsOptions struct {
	Container    string    `qs:"-"`
	OutputStream io.Writer `qs:"-"`
	ErrorStream  io.Writer `qs:"-"`
	Follow       bool
	Stdout       bool
	Stderr       bool
	Timestamps   bool
	Tail         string

	// Use raw terminal? Usually true when the container contains a TTY.
	RawTerminal bool `qs:"-"`
}

// ExportContainerOptions is the set of parameters to the ExportContainer
// method.
//
// See http://goo.gl/hnzE62 for more details.
type ExportContainerOptions struct {
	ID           string
	OutputStream io.Writer
}

// NoSuchContainer is the error returned when a given container does not exist.
type NoSuchContainer struct {
	ID  string
	Err error
}

// ContainerAlreadyRunning is the error returned when a given container is
// already running.
type ContainerAlreadyRunning struct {
	ID string
}

// ContainerNotRunning is the error returned when a given container is not
// running.
type ContainerNotRunning struct {
	ID string
}

//****************************************************************//
//env use type
//****************************************************************//

// Env represents a list of key-pair represented in the form KEY=VALUE.
type Env []string

//****************************************************************//
//enent use type
//****************************************************************//

// APIEvents represents an event returned by the API.
type APIEvents struct {
	Status string `json:"Status,omitempty" yaml:"Status,omitempty"`
	ID     string `json:"ID,omitempty" yaml:"ID,omitempty"`
	From   string `json:"From,omitempty" yaml:"From,omitempty"`
	Time   int64  `json:"Time,omitempty" yaml:"Time,omitempty"`
}

type eventMonitoringState struct {
	sync.RWMutex
	sync.WaitGroup
	enabled   bool
	lastSeen  *int64
	C         chan *APIEvents
	errC      chan error
	listeners []chan<- *APIEvents
}

//****************************************************************//
//exec use type
//****************************************************************//

// CreateExecOptions specify parameters to the CreateExecContainer function.
//
// See http://goo.gl/8izrzI for more details
type CreateExecOptions struct {
	AttachStdin  bool     `json:"AttachStdin,omitempty" yaml:"AttachStdin,omitempty"`
	AttachStdout bool     `json:"AttachStdout,omitempty" yaml:"AttachStdout,omitempty"`
	AttachStderr bool     `json:"AttachStderr,omitempty" yaml:"AttachStderr,omitempty"`
	Tty          bool     `json:"Tty,omitempty" yaml:"Tty,omitempty"`
	Cmd          []string `json:"Cmd,omitempty" yaml:"Cmd,omitempty"`
	Container    string   `json:"Container,omitempty" yaml:"Container,omitempty"`
	User         string   `json:"User,omitempty" yaml:"User,omitempty"`
}

// StartExecOptions specify parameters to the StartExecContainer function.
//
// See http://goo.gl/JW8Lxl for more details
type StartExecOptions struct {
	Detach bool `json:"Detach,omitempty" yaml:"Detach,omitempty"`

	Tty bool `json:"Tty,omitempty" yaml:"Tty,omitempty"`

	InputStream  io.Reader `qs:"-"`
	OutputStream io.Writer `qs:"-"`
	ErrorStream  io.Writer `qs:"-"`

	// Use raw terminal? Usually true when the container contains a TTY.
	RawTerminal bool `qs:"-"`

	// If set, after a successful connect, a sentinel will be sent and then the
	// client will block on receive before continuing.
	//
	// It must be an unbuffered channel. Using a buffered channel can lead
	// to unexpected behavior.
	Success chan struct{} `json:"-"`
}

// Exec is the type representing a `docker exec` instance and containing the
// instance ID
type Exec struct {
	ID string `json:"Id,omitempty" yaml:"Id,omitempty"`
}

// ExecProcessConfig is a type describing the command associated to a Exec
// instance. It's used in the ExecInspect type.
//
// See http://goo.gl/ypQULN for more details
type ExecProcessConfig struct {
	Privileged bool     `json:"privileged,omitempty" yaml:"privileged,omitempty"`
	User       string   `json:"user,omitempty" yaml:"user,omitempty"`
	Tty        bool     `json:"tty,omitempty" yaml:"tty,omitempty"`
	EntryPoint string   `json:"entrypoint,omitempty" yaml:"entrypoint,omitempty"`
	Arguments  []string `json:"arguments,omitempty" yaml:"arguments,omitempty"`
}

// ExecInspect is a type with details about a exec instance, including the
// exit code if the command has finished running. It's returned by a api
// call to /exec/(id)/json
//
// See http://goo.gl/ypQULN for more details
type ExecInspect struct {
	ID            string            `json:"ID,omitempty" yaml:"ID,omitempty"`
	Running       bool              `json:"Running,omitempty" yaml:"Running,omitempty"`
	ExitCode      int               `json:"ExitCode,omitempty" yaml:"ExitCode,omitempty"`
	OpenStdin     bool              `json:"OpenStdin,omitempty" yaml:"OpenStdin,omitempty"`
	OpenStderr    bool              `json:"OpenStderr,omitempty" yaml:"OpenStderr,omitempty"`
	OpenStdout    bool              `json:"OpenStdout,omitempty" yaml:"OpenStdout,omitempty"`
	ProcessConfig ExecProcessConfig `json:"ProcessConfig,omitempty" yaml:"ProcessConfig,omitempty"`
	Container     Container         `json:"Container,omitempty" yaml:"Container,omitempty"`
}

// NoSuchExec is the error returned when a given exec instance does not exist.
type NoSuchExec struct {
	ID string
}

//****************************************************************//
//image use type
//****************************************************************//

// APIImages represent an image returned in the ListImages call.
type APIImages struct {
	ID          string   `json:"Id" yaml:"Id"`
	RepoTags    []string `json:"RepoTags,omitempty" yaml:"RepoTags,omitempty"`
	Created     int64    `json:"Created,omitempty" yaml:"Created,omitempty"`
	Size        int64    `json:"Size,omitempty" yaml:"Size,omitempty"`
	VirtualSize int64    `json:"VirtualSize,omitempty" yaml:"VirtualSize,omitempty"`
	ParentID    string   `json:"ParentId,omitempty" yaml:"ParentId,omitempty"`
	RepoDigests []string `json:"RepoDigests,omitempty" yaml:"RepoDigests,omitempty"`
}

// Image is the type representing a docker image and its various properties
type Image struct {
	ID              string    `json:"Id" yaml:"Id"`
	Parent          string    `json:"Parent,omitempty" yaml:"Parent,omitempty"`
	Comment         string    `json:"Comment,omitempty" yaml:"Comment,omitempty"`
	Created         time.Time `json:"Created,omitempty" yaml:"Created,omitempty"`
	Container       string    `json:"Container,omitempty" yaml:"Container,omitempty"`
	ContainerConfig Config    `json:"ContainerConfig,omitempty" yaml:"ContainerConfig,omitempty"`
	DockerVersion   string    `json:"DockerVersion,omitempty" yaml:"DockerVersion,omitempty"`
	Author          string    `json:"Author,omitempty" yaml:"Author,omitempty"`
	Config          *Config   `json:"Config,omitempty" yaml:"Config,omitempty"`
	Architecture    string    `json:"Architecture,omitempty" yaml:"Architecture,omitempty"`
	Size            int64     `json:"Size,omitempty" yaml:"Size,omitempty"`
	VirtualSize     int64     `json:"VirtualSize,omitempty" yaml:"VirtualSize,omitempty"`
}

// ImageHistory represent a layer in an image's history returned by the
// ImageHistory call.
type ImageHistory struct {
	ID        string   `json:"Id" yaml:"Id"`
	Tags      []string `json:"Tags,omitempty" yaml:"Tags,omitempty"`
	Created   int64    `json:"Created,omitempty" yaml:"Created,omitempty"`
	CreatedBy string   `json:"CreatedBy,omitempty" yaml:"CreatedBy,omitempty"`
	Size      int64    `json:"Size,omitempty" yaml:"Size,omitempty"`
}

// ImagePre012 serves the same purpose as the Image type except that it is for
// earlier versions of the Docker API (pre-012 to be specific)
type ImagePre012 struct {
	ID              string    `json:"id"`
	Parent          string    `json:"parent,omitempty"`
	Comment         string    `json:"comment,omitempty"`
	Created         time.Time `json:"created"`
	Container       string    `json:"container,omitempty"`
	ContainerConfig Config    `json:"container_config,omitempty"`
	DockerVersion   string    `json:"docker_version,omitempty"`
	Author          string    `json:"author,omitempty"`
	Config          *Config   `json:"config,omitempty"`
	Architecture    string    `json:"architecture,omitempty"`
	Size            int64     `json:"size,omitempty"`
}

// ListImagesOptions specify parameters to the ListImages function.
//
// See http://goo.gl/HRVN1Z for more details.
type ListImagesOptions struct {
	All     bool
	Filters map[string][]string
	Digests bool
}

// RemoveImageOptions present the set of options available for removing an image
// from a registry.
//
// See http://goo.gl/6V48bF for more details.
type RemoveImageOptions struct {
	Force   bool `qs:"force"`
	NoPrune bool `qs:"noprune"`
}

// PushImageOptions represents options to use in the PushImage method.
//
// See http://goo.gl/pN8A3P for more details.
type PushImageOptions struct {
	// Name of the image
	Name string

	// Tag of the image
	Tag string

	// Registry server to push the image
	Registry string

	OutputStream  io.Writer `qs:"-"`
	RawJSONStream bool      `qs:"-"`
}

// PullImageOptions present the set of options available for pulling an image
// from a registry.
//
// See http://goo.gl/ACyYNS for more details.
type PullImageOptions struct {
	Repository    string `qs:"fromImage"`
	Registry      string
	Tag           string
	OutputStream  io.Writer `qs:"-"`
	RawJSONStream bool      `qs:"-"`
}

// LoadImageOptions represents the options for LoadImage Docker API Call
//
// See http://goo.gl/Y8NNCq for more details.
type LoadImageOptions struct {
	InputStream io.Reader
}

// ExportImageOptions represent the options for ExportImage Docker API call
//
// See http://goo.gl/mi6kvk for more details.
type ExportImageOptions struct {
	Name         string
	OutputStream io.Writer
}

// ExportImagesOptions represent the options for ExportImages Docker API call
//
// See http://goo.gl/YeZzQK for more details.
type ExportImagesOptions struct {
	Names        []string
	OutputStream io.Writer `qs:"-"`
}

// ImportImageOptions present the set of informations available for importing
// an image from a source file or the stdin.
//
// See http://goo.gl/PhBKnS for more details.
type ImportImageOptions struct {
	Repository string `qs:"repo"`
	Source     string `qs:"fromSrc"`
	Tag        string `qs:"tag"`

	InputStream   io.Reader `qs:"-"`
	OutputStream  io.Writer `qs:"-"`
	RawJSONStream bool      `qs:"-"`
}

// BuildImageOptions present the set of informations available for building an
// image from a tarfile with a Dockerfile in it.
//
// For more details about the Docker building process, see
// http://goo.gl/tlPXPu.
type BuildImageOptions struct {
	Name                string             `qs:"t"`
	Dockerfile          string             `qs:"dockerfile"`
	NoCache             bool               `qs:"nocache"`
	SuppressOutput      bool               `qs:"q"`
	Pull                bool               `qs:"pull"`
	RmTmpContainer      bool               `qs:"rm"`
	ForceRmTmpContainer bool               `qs:"forcerm"`
	Memory              int64              `qs:"memory"`
	Memswap             int64              `qs:"memswap"`
	CPUShares           int64              `qs:"cpushares"`
	CPUSetCPUs          string             `qs:"cpusetcpus"`
	InputStream         io.Reader          `qs:"-"`
	OutputStream        io.Writer          `qs:"-"`
	RawJSONStream       bool               `qs:"-"`
	Remote              string             `qs:"remote"`
	Auth                AuthConfiguration  `qs:"-"` // for older docker X-Registry-Auth header
	AuthConfigs         AuthConfigurations `qs:"-"` // for newer docker X-Registry-Config header
	ContextDir          string             `qs:"-"`
}

// TagImageOptions present the set of options to tag an image.
//
// See http://goo.gl/5g6qFy for more details.
type TagImageOptions struct {
	Repo  string
	Tag   string
	Force bool
}

// APIImageSearch reflect the result of a search on the dockerHub
//
// See http://goo.gl/xI5lLZ for more details.
type APIImageSearch struct {
	Description string `json:"description,omitempty" yaml:"description,omitempty"`
	IsOfficial  bool   `json:"is_official,omitempty" yaml:"is_official,omitempty"`
	IsAutomated bool   `json:"is_automated,omitempty" yaml:"is_automated,omitempty"`
	Name        string `json:"name,omitempty" yaml:"name,omitempty"`
	StarCount   int    `json:"star_count,omitempty" yaml:"star_count,omitempty"`
}

//****************************************************************//
//signal use type
//****************************************************************//

// Signal represents a signal that can be send to the container on
// KillContainer call.
type Signal int

//****************************************************************//
//tls use type
//****************************************************************//

type tlsClientCon struct {
	*tls.Conn
	rawConn net.Conn
}

//****************************************************************//
//change use const
//****************************************************************//

const (
	// ChangeModify is the ChangeType for container modifications
	ChangeModify ChangeType = iota

	// ChangeAdd is the ChangeType for additions to a container
	ChangeAdd

	// ChangeDelete is the ChangeType for deletions from a container
	ChangeDelete
)

//****************************************************************//
//client use const
//****************************************************************//

const generatorAgent = "go-generator-client"

//****************************************************************//
//event use const
//****************************************************************//

const (
	maxMonitorConnRetries = 5
	retryInitialWaitTime  = 10.
)

//****************************************************************//
//signal use const
//****************************************************************//

// These values represent all signals available on Linux, where containers will
// be running.
const (
	SIGABRT   = Signal(0x6)
	SIGALRM   = Signal(0xe)
	SIGBUS    = Signal(0x7)
	SIGCHLD   = Signal(0x11)
	SIGCLD    = Signal(0x11)
	SIGCONT   = Signal(0x12)
	SIGFPE    = Signal(0x8)
	SIGHUP    = Signal(0x1)
	SIGILL    = Signal(0x4)
	SIGINT    = Signal(0x2)
	SIGIO     = Signal(0x1d)
	SIGIOT    = Signal(0x6)
	SIGKILL   = Signal(0x9)
	SIGPIPE   = Signal(0xd)
	SIGPOLL   = Signal(0x1d)
	SIGPROF   = Signal(0x1b)
	SIGPWR    = Signal(0x1e)
	SIGQUIT   = Signal(0x3)
	SIGSEGV   = Signal(0xb)
	SIGSTKFLT = Signal(0x10)
	SIGSTOP   = Signal(0x13)
	SIGSYS    = Signal(0x1f)
	SIGTERM   = Signal(0xf)
	SIGTRAP   = Signal(0x5)
	SIGTSTP   = Signal(0x14)
	SIGTTIN   = Signal(0x15)
	SIGTTOU   = Signal(0x16)
	SIGUNUSED = Signal(0x1f)
	SIGURG    = Signal(0x17)
	SIGUSR1   = Signal(0xa)
	SIGUSR2   = Signal(0xc)
	SIGVTALRM = Signal(0x1a)
	SIGWINCH  = Signal(0x1c)
	SIGXCPU   = Signal(0x18)
	SIGXFSZ   = Signal(0x19)
)

//****************************************************************//
//client use var
//****************************************************************//

var (
	// ErrInvalidEndpoint is returned when the endpoint is not a valid HTTP URL.
	ErrInvalidEndpoint = errors.New("invalid endpoint")

	// ErrConnectionRefused is returned when the client cannot connect to the given endpoint.
	ErrConnectionRefused = errors.New("cannot connect to Docker endpoint")

	apiVersion112, _ = NewAPIVersion("1.12")
)

//****************************************************************//
//Container use var
//****************************************************************//

// ErrContainerAlreadyExists is the error returned by CreateContainer when the container already exists.
var ErrContainerAlreadyExists = errors.New("container already exists")

//****************************************************************//
//event use var
//****************************************************************//

var (
	// ErrNoListeners is the error returned when no listeners are available
	// to receive an event.
	ErrNoListeners = errors.New("no listeners present to receive event")

	// ErrListenerAlreadyExists is the error returned when the listerner already
	// exists.
	ErrListenerAlreadyExists = errors.New("listener already exists for docker events")

	// EOFEvent is sent when the event listener receives an EOF error.
	EOFEvent = &APIEvents{
		Status: "EOF",
	}
)

//****************************************************************//
//image use var
//****************************************************************//

var (
	// ErrNoSuchImage is the error returned when the image does not exist.
	ErrNoSuchImage = errors.New("no such image")

	// ErrMissingRepo is the error returned when the remote repository is
	// missing.
	ErrMissingRepo = errors.New("missing remote repository e.g. 'github.com/user/repo'")

	// ErrMissingOutputStream is the error returned when no output stream
	// is provided to some calls, like BuildImage.
	ErrMissingOutputStream = errors.New("missing output stream")

	// ErrMultipleContexts is the error returned when both a ContextDir and
	// InputStream are provided in BuildImageOptions
	ErrMultipleContexts = errors.New("image build may not be provided BOTH context dir and input stream")

	// ErrMustSpecifyNames is the error rreturned when the Names field on
	// ExportImagesOptions is nil or empty
	ErrMustSpecifyNames = errors.New("must specify at least one name to export")
)

//****************************************************************//
//auth need func
//****************************************************************//

// NewAuthConfigurationsFromDockerCfg returns AuthConfigurations from the
// ~/.dockercfg file.
func NewAuthConfigurationsFromDockerCfg() (*AuthConfigurations, error) {
	p := path.Join(os.Getenv("HOME"), ".dockercfg")
	r, err := os.Open(p)
	if err != nil {
		return nil, err
	}
	return NewAuthConfigurations(r)
}

// NewAuthConfigurations returns AuthConfigurations from a JSON encoded string in the
// same format as the .dockercfg file.
func NewAuthConfigurations(r io.Reader) (*AuthConfigurations, error) {
	var auth *AuthConfigurations
	var confs map[string]dockerConfig
	if err := json.NewDecoder(r).Decode(&confs); err != nil {
		return nil, err
	}
	auth, err := authConfigs(confs)
	if err != nil {
		return nil, err
	}
	return auth, nil
}

// authConfigs converts a dockerConfigs map to a AuthConfigurations object.
func authConfigs(confs map[string]dockerConfig) (*AuthConfigurations, error) {
	c := &AuthConfigurations{
		Configs: make(map[string]AuthConfiguration),
	}
	for reg, conf := range confs {
		data, err := base64.StdEncoding.DecodeString(conf.Auth)
		if err != nil {
			return nil, err
		}
		userpass := strings.Split(string(data), ":")
		c.Configs[reg] = AuthConfiguration{
			Email:         conf.Email,
			Username:      userpass[0],
			Password:      userpass[1],
			ServerAddress: reg,
		}
	}
	return c, nil
}

// AuthCheck validates the given credentials. It returns nil if successful.
//
// See https://goo.gl/vPoEfJ for more details.
func (c *Client) AuthCheck(conf *AuthConfiguration) error {
	if conf == nil {
		return fmt.Errorf("conf is nil")
	}
	body, statusCode, err := c.do("POST", "/auth", doOptions{data: conf})
	if err != nil {
		return err
	}
	if statusCode > 400 {
		return fmt.Errorf("auth error (%d): %s", statusCode, body)
	}
	return nil
}

//****************************************************************//
//change need func
//****************************************************************//

func (change *Change) String() string {
	var kind string
	switch change.Kind {
	case ChangeModify:
		kind = "C"
	case ChangeAdd:
		kind = "A"
	case ChangeDelete:
		kind = "D"
	}
	return fmt.Sprintf("%s %s", kind, change.Path)
}

//****************************************************************//
//client need func
//****************************************************************//

// NewAPIVersion returns an instance of APIVersion for the given string.
//
// The given string must be in the form <major>.<minor>.<patch>, where <major>,
// <minor> and <patch> are integer numbers.
func NewAPIVersion(input string) (APIVersion, error) {
	if !strings.Contains(input, ".") {
		return nil, fmt.Errorf("Unable to parse version %q", input)
	}
	arr := strings.Split(input, ".")
	ret := make(APIVersion, len(arr))
	var err error
	for i, val := range arr {
		ret[i], err = strconv.Atoi(val)
		if err != nil {
			return nil, fmt.Errorf("Unable to parse version %q: %q is not an integer", input, val)
		}
	}
	return ret, nil
}

func (version APIVersion) String() string {
	var str string
	for i, val := range version {
		str += strconv.Itoa(val)
		if i < len(version)-1 {
			str += "."
		}
	}
	return str
}

// LessThan is a function for comparing APIVersion structs
func (version APIVersion) LessThan(other APIVersion) bool {
	return version.compare(other) < 0
}

// LessThanOrEqualTo is a function for comparing APIVersion structs
func (version APIVersion) LessThanOrEqualTo(other APIVersion) bool {
	return version.compare(other) <= 0
}

// GreaterThan is a function for comparing APIVersion structs
func (version APIVersion) GreaterThan(other APIVersion) bool {
	return version.compare(other) > 0
}

// GreaterThanOrEqualTo is a function for comparing APIVersion structs
func (version APIVersion) GreaterThanOrEqualTo(other APIVersion) bool {
	return version.compare(other) >= 0
}

func (version APIVersion) compare(other APIVersion) int {
	for i, v := range version {
		if i <= len(other)-1 {
			otherVersion := other[i]

			if v < otherVersion {
				return -1
			} else if v > otherVersion {
				return 1
			}
		}
	}
	if len(version) > len(other) {
		return 1
	}
	if len(version) < len(other) {
		return -1
	}
	return 0
}

// NewClient returns a Client instance ready for communication with the given
// server endpoint. It will use the latest remote API version available in the
// server.
func NewClient(endpoint string) (*Client, error) {
	client, err := NewVersionedClient(endpoint, "")
	if err != nil {
		return nil, err
	}
	client.SkipServerVersionCheck = true
	return client, nil
}

// NewTLSClient returns a Client instance ready for TLS communications with the givens
// server endpoint, key and certificates . It will use the latest remote API version
// available in the server.
func NewTLSClient(endpoint string, cert, key, ca string) (*Client, error) {
	client, err := NewVersionedTLSClient(endpoint, cert, key, ca, "")
	if err != nil {
		return nil, err
	}
	client.SkipServerVersionCheck = true
	return client, nil
}

// NewTLSClientFromBytes returns a Client instance ready for TLS communications with the givens
// server endpoint, key and certificates (passed inline to the function as opposed to being
// read from a local file). It will use the latest remote API version available in the server.
func NewTLSClientFromBytes(endpoint string, certPEMBlock, keyPEMBlock, caPEMCert []byte) (*Client, error) {
	client, err := NewVersionedTLSClientFromBytes(endpoint, certPEMBlock, keyPEMBlock, caPEMCert, "")
	if err != nil {
		return nil, err
	}
	client.SkipServerVersionCheck = true
	return client, nil
}

// NewVersionedClient returns a Client instance ready for communication with
// the given server endpoint, using a specific remote API version.
func NewVersionedClient(endpoint string, apiVersionString string) (*Client, error) {
	u, err := parseEndpoint(endpoint, false)
	if err != nil {
		return nil, err
	}
	var requestedAPIVersion APIVersion
	if strings.Contains(apiVersionString, ".") {
		requestedAPIVersion, err = NewAPIVersion(apiVersionString)
		if err != nil {
			return nil, err
		}
	}

	tr := &http.Transport{}
	return &Client{
		HTTPClient:          http.DefaultClient,
		transport:           tr,
		endpoint:            endpoint,
		endpointURL:         u,
		eventMonitor:        new(eventMonitoringState),
		requestedAPIVersion: requestedAPIVersion,
	}, nil
}

// NewVersionnedTLSClient has been DEPRECATED, please use NewVersionedTLSClient.
func NewVersionnedTLSClient(endpoint string, cert, key, ca, apiVersionString string) (*Client, error) {
	return NewVersionedTLSClient(endpoint, cert, key, ca, apiVersionString)
}

// NewVersionedTLSClient returns a Client instance ready for TLS communications with the givens
// server endpoint, key and certificates, using a specific remote API version.
func NewVersionedTLSClient(endpoint string, cert, key, ca, apiVersionString string) (*Client, error) {
	certPEMBlock, err := ioutil.ReadFile(cert)
	if err != nil {
		return nil, err
	}
	keyPEMBlock, err := ioutil.ReadFile(key)
	if err != nil {
		return nil, err
	}
	caPEMCert, err := ioutil.ReadFile(ca)
	if err != nil {
		return nil, err
	}
	return NewVersionedTLSClientFromBytes(endpoint, certPEMBlock, keyPEMBlock, caPEMCert, apiVersionString)
}

// NewVersionedTLSClientFromBytes returns a Client instance ready for TLS communications with the givens
// server endpoint, key and certificates (passed inline to the function as opposed to being
// read from a local file), using a specific remote API version.
func NewVersionedTLSClientFromBytes(endpoint string, certPEMBlock, keyPEMBlock, caPEMCert []byte, apiVersionString string) (*Client, error) {
	u, err := parseEndpoint(endpoint, true)
	if err != nil {
		return nil, err
	}
	var requestedAPIVersion APIVersion
	if strings.Contains(apiVersionString, ".") {
		requestedAPIVersion, err = NewAPIVersion(apiVersionString)
		if err != nil {
			return nil, err
		}
	}
	if certPEMBlock == nil || keyPEMBlock == nil {
		return nil, errors.New("Both cert and key are required")
	}
	tlsCert, err := tls.X509KeyPair(certPEMBlock, keyPEMBlock)
	if err != nil {
		return nil, err
	}
	tlsConfig := &tls.Config{Certificates: []tls.Certificate{tlsCert}}
	if caPEMCert == nil {
		tlsConfig.InsecureSkipVerify = true
	} else {
		caPool := x509.NewCertPool()
		if !caPool.AppendCertsFromPEM(caPEMCert) {
			return nil, errors.New("Could not add RootCA pem")
		}
		tlsConfig.RootCAs = caPool
	}
	tr := &http.Transport{
		TLSClientConfig: tlsConfig,
	}
	if err != nil {
		return nil, err
	}
	return &Client{
		HTTPClient:          &http.Client{Transport: tr},
		transport:           tr,
		TLSConfig:           tlsConfig,
		endpoint:            endpoint,
		endpointURL:         u,
		eventMonitor:        new(eventMonitoringState),
		requestedAPIVersion: requestedAPIVersion,
	}, nil
}

func (c *Client) checkAPIVersion() error {
	serverAPIVersionString, err := c.getServerAPIVersionString()
	if err != nil {
		return err
	}
	c.serverAPIVersion, err = NewAPIVersion(serverAPIVersionString)
	if err != nil {
		return err
	}
	if c.requestedAPIVersion == nil {
		c.expectedAPIVersion = c.serverAPIVersion
	} else {
		c.expectedAPIVersion = c.requestedAPIVersion
	}
	return nil
}

// Ping pings the docker server
//
// See http://goo.gl/stJENm for more details.
func (c *Client) Ping() error {
	path := "/_ping"
	body, status, err := c.do("GET", path, doOptions{})
	if err != nil {
		return err
	}
	if status != http.StatusOK {
		return newError(status, body)
	}
	return nil
}

func (c *Client) getServerAPIVersionString() (version string, err error) {
	body, status, err := c.do("GET", "/version", doOptions{})
	if err != nil {
		return "", err
	}
	if status != http.StatusOK {
		return "", fmt.Errorf("Received unexpected status %d while trying to retrieve the server version", status)
	}
	var versionResponse map[string]interface{}
	err = json.Unmarshal(body, &versionResponse)
	if err != nil {
		return "", err
	}
	if version, ok := (versionResponse["ApiVersion"]).(string); ok {
		return version, nil
	}
	return "", nil
}

func (c *Client) do(method, path string, doOptions doOptions) ([]byte, int, error) {
	var params io.Reader
	if doOptions.data != nil || doOptions.forceJSON {
		buf, err := json.Marshal(doOptions.data)
		if err != nil {
			return nil, -1, err
		}
		params = bytes.NewBuffer(buf)
	}
	if path != "/version" && !c.SkipServerVersionCheck && c.expectedAPIVersion == nil {
		err := c.checkAPIVersion()
		if err != nil {
			return nil, -1, err
		}
	}
	req, err := http.NewRequest(method, c.getURL(path), params)
	if err != nil {
		return nil, -1, err
	}
	req.Header.Set("User-Agent", generatorAgent)
	if doOptions.data != nil {
		req.Header.Set("Content-Type", "application/json")
	} else if method == "POST" {
		req.Header.Set("Content-Type", "plain/text")
	}
	var resp *http.Response
	protocol := c.endpointURL.Scheme
	address := c.endpointURL.Path
	if protocol == "unix" {
		dial, err := net.Dial(protocol, address)
		if err != nil {
			return nil, -1, err
		}
		defer dial.Close()
		breader := bufio.NewReader(dial)
		err = req.Write(dial)
		if err != nil {
			return nil, -1, err
		}
		resp, err = http.ReadResponse(breader, req)
	} else {
		resp, err = c.HTTPClient.Do(req)
	}
	if err != nil {
		if strings.Contains(err.Error(), "connection refused") {
			return nil, -1, ErrConnectionRefused
		}
		return nil, -1, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, -1, err
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		return nil, resp.StatusCode, newError(resp.StatusCode, body)
	}
	return body, resp.StatusCode, nil
}

func (c *Client) stream(method, path string, streamOptions streamOptions) error {
	if (method == "POST" || method == "PUT") && streamOptions.in == nil {
		streamOptions.in = bytes.NewReader(nil)
	}
	if path != "/version" && !c.SkipServerVersionCheck && c.expectedAPIVersion == nil {
		err := c.checkAPIVersion()
		if err != nil {
			return err
		}
	}
	req, err := http.NewRequest(method, c.getURL(path), streamOptions.in)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", generatorAgent)
	if method == "POST" {
		req.Header.Set("Content-Type", "plain/text")
	}
	for key, val := range streamOptions.headers {
		req.Header.Set(key, val)
	}
	var resp *http.Response
	protocol := c.endpointURL.Scheme
	address := c.endpointURL.Path
	if streamOptions.stdout == nil {
		streamOptions.stdout = ioutil.Discard
	}
	if streamOptions.stderr == nil {
		streamOptions.stderr = ioutil.Discard
	}
	if protocol == "unix" {
		dial, err := net.Dial(protocol, address)
		if err != nil {
			return err
		}
		defer dial.Close()
		breader := bufio.NewReader(dial)
		err = req.Write(dial)
		if err != nil {
			return err
		}
		if resp, err = http.ReadResponse(breader, req); err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				return ErrConnectionRefused
			}
			return err
		}
		defer resp.Body.Close()
	} else {
		if resp, err = c.HTTPClient.Do(req); err != nil {
			if strings.Contains(err.Error(), "connection refused") {
				return ErrConnectionRefused
			}
			return err
		}
		defer resp.Body.Close()
		defer c.transport.CancelRequest(req)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 400 {
		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return err
		}
		return newError(resp.StatusCode, body)
	}
	if streamOptions.useJSONDecoder || resp.Header.Get("Content-Type") == "application/json" {
		// if we want to get raw json stream, just copy it back to output
		// without decoding it
		if streamOptions.rawJSONStream {
			_, err = io.Copy(streamOptions.stdout, resp.Body)
			return err
		}
		dec := json.NewDecoder(resp.Body)
		for {
			var m jsonMessage
			if err := dec.Decode(&m); err == io.EOF {
				break
			} else if err != nil {
				return err
			}
			if m.Stream != "" {
				fmt.Fprint(streamOptions.stdout, m.Stream)
			} else if m.Progress != "" {
				fmt.Fprintf(streamOptions.stdout, "%s %s\r", m.Status, m.Progress)
			} else if m.Error != "" {
				return errors.New(m.Error)
			}
			if m.Status != "" {
				fmt.Fprintln(streamOptions.stdout, m.Status)
			}
		}
	} else {
		if streamOptions.setRawTerminal {
			_, err = io.Copy(streamOptions.stdout, resp.Body)
		} else {
			_, err = stdcopy.StdCopy(streamOptions.stdout, streamOptions.stderr, resp.Body)
		}
		return err
	}
	return nil
}

func (c *Client) hijack(method, path string, hijackOptions hijackOptions) error {
	if path != "/version" && !c.SkipServerVersionCheck && c.expectedAPIVersion == nil {
		err := c.checkAPIVersion()
		if err != nil {
			return err
		}
	}

	var params io.Reader
	if hijackOptions.data != nil {
		buf, err := json.Marshal(hijackOptions.data)
		if err != nil {
			return err
		}
		params = bytes.NewBuffer(buf)
	}

	if hijackOptions.stdout == nil {
		hijackOptions.stdout = ioutil.Discard
	}
	if hijackOptions.stderr == nil {
		hijackOptions.stderr = ioutil.Discard
	}
	req, err := http.NewRequest(method, c.getURL(path), params)
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "plain/text")
	protocol := c.endpointURL.Scheme
	address := c.endpointURL.Path
	if protocol != "unix" {
		protocol = "tcp"
		address = c.endpointURL.Host
	}
	var dial net.Conn
	if c.TLSConfig != nil && protocol != "unix" {
		dial, err = tlsDial(protocol, address, c.TLSConfig)
		if err != nil {
			return err
		}
	} else {
		dial, err = net.Dial(protocol, address)
		if err != nil {
			return err
		}
	}
	clientconn := httputil.NewClientConn(dial, nil)
	defer clientconn.Close()
	clientconn.Do(req)
	if hijackOptions.success != nil {
		hijackOptions.success <- struct{}{}
		<-hijackOptions.success
	}
	rwc, br := clientconn.Hijack()
	defer rwc.Close()
	errChanOut := make(chan error, 1)
	errChanIn := make(chan error, 1)
	exit := make(chan bool)
	go func() {
		defer close(exit)
		defer close(errChanOut)
		var err error
		if hijackOptions.setRawTerminal {
			// When TTY is ON, use regular copy
			_, err = io.Copy(hijackOptions.stdout, br)
		} else {
			_, err = stdcopy.StdCopy(hijackOptions.stdout, hijackOptions.stderr, br)
		}
		errChanOut <- err
	}()
	go func() {
		if hijackOptions.in != nil {
			_, err := io.Copy(rwc, hijackOptions.in)
			errChanIn <- err
		}
		rwc.(interface {
			CloseWrite() error
		}).CloseWrite()
	}()
	<-exit
	select {
	case err = <-errChanIn:
		return err
	case err = <-errChanOut:
		return err
	}
}

func (c *Client) getURL(path string) string {
	urlStr := strings.TrimRight(c.endpointURL.String(), "/")
	if c.endpointURL.Scheme == "unix" {
		urlStr = ""
	}

	if c.requestedAPIVersion != nil {
		return fmt.Sprintf("%s/v%s%s", urlStr, c.requestedAPIVersion, path)
	}
	return fmt.Sprintf("%s%s", urlStr, path)
}

func queryString(opts interface{}) string {
	if opts == nil {
		return ""
	}
	value := reflect.ValueOf(opts)
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}
	if value.Kind() != reflect.Struct {
		return ""
	}
	items := url.Values(map[string][]string{})
	for i := 0; i < value.NumField(); i++ {
		field := value.Type().Field(i)
		if field.PkgPath != "" {
			continue
		}
		key := field.Tag.Get("qs")
		if key == "" {
			key = strings.ToLower(field.Name)
		} else if key == "-" {
			continue
		}
		addQueryStringValue(items, key, value.Field(i))
	}
	return items.Encode()
}

func addQueryStringValue(items url.Values, key string, v reflect.Value) {
	switch v.Kind() {
	case reflect.Bool:
		if v.Bool() {
			items.Add(key, "1")
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if v.Int() > 0 {
			items.Add(key, strconv.FormatInt(v.Int(), 10))
		}
	case reflect.Float32, reflect.Float64:
		if v.Float() > 0 {
			items.Add(key, strconv.FormatFloat(v.Float(), 'f', -1, 64))
		}
	case reflect.String:
		if v.String() != "" {
			items.Add(key, v.String())
		}
	case reflect.Ptr:
		if !v.IsNil() {
			if b, err := json.Marshal(v.Interface()); err == nil {
				items.Add(key, string(b))
			}
		}
	case reflect.Map:
		if len(v.MapKeys()) > 0 {
			if b, err := json.Marshal(v.Interface()); err == nil {
				items.Add(key, string(b))
			}
		}
	case reflect.Array, reflect.Slice:
		vLen := v.Len()
		if vLen > 0 {
			for i := 0; i < vLen; i++ {
				addQueryStringValue(items, key, v.Index(i))
			}
		}
	}
}

func newError(status int, body []byte) *Error {
	return &Error{Status: status, Message: string(body)}
}

func (e *Error) Error() string {
	return fmt.Sprintf("API error (%d): %s", e.Status, e.Message)
}

func parseEndpoint(endpoint string, tls bool) (*url.URL, error) {
	u, err := url.Parse(endpoint)
	if err != nil {
		return nil, ErrInvalidEndpoint
	}
	if tls {
		u.Scheme = "https"
	}
	switch u.Scheme {
	case "unix":
		return u, nil
	case "http", "https", "tcp":
		_, port, err := net.SplitHostPort(u.Host)
		if err != nil {
			if e, ok := err.(*net.AddrError); ok {
				if e.Err == "missing port in address" {
					return u, nil
				}
			}
			return nil, ErrInvalidEndpoint
		}
		number, err := strconv.ParseInt(port, 10, 64)
		if err == nil && number > 0 && number < 65536 {
			if u.Scheme == "tcp" {
				if number == 2376 {
					u.Scheme = "https"
				} else {
					u.Scheme = "http"
				}
			}
			return u, nil
		}
		return nil, ErrInvalidEndpoint
	default:
		return nil, ErrInvalidEndpoint
	}
}

//****************************************************************//
//container need func
//****************************************************************//

// ListContainers returns a slice of containers matching the given criteria.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#list-containers for more details.
func (c *Client) ListContainers(opts ListContainersOptions) ([]APIContainers, error) {
	path := "/containers/json?" + queryString(opts)
	body, _, err := c.do("GET", path, doOptions{})
	if err != nil {
		return nil, err
	}
	var containers []APIContainers
	err = json.Unmarshal(body, &containers)
	if err != nil {
		return nil, err
	}
	return containers, nil
}

// Port returns the number of the port.
func (p Port) Port() string {
	return strings.Split(string(p), "/")[0]
}

// Proto returns the name of the protocol.
func (p Port) Proto() string {
	parts := strings.Split(string(p), "/")
	if len(parts) == 1 {
		return "tcp"
	}
	return parts[1]
}

// String returns the string representation of a state.
func (s *State) String() string {
	if s.Running {
		if s.Paused {
			return "paused"
		}
		return fmt.Sprintf("Up %s", time.Now().UTC().Sub(s.StartedAt))
	}
	return fmt.Sprintf("Exit %d", s.ExitCode)
}

// PortMappingAPI translates the port mappings as contained in NetworkSettings
// into the format in which they would appear when returned by the API
func (settings *NetworkSettings) PortMappingAPI() []APIPort {
	var mapping []APIPort
	for port, bindings := range settings.Ports {
		p, _ := parsePort(port.Port())
		if len(bindings) == 0 {
			mapping = append(mapping, APIPort{
				PublicPort: int64(p),
				Type:       port.Proto(),
			})
			continue
		}
		for _, binding := range bindings {
			p, _ := parsePort(port.Port())
			h, _ := parsePort(binding.HostPort)
			mapping = append(mapping, APIPort{
				PrivatePort: int64(p),
				PublicPort:  int64(h),
				Type:        port.Proto(),
				IP:          binding.HostIP,
			})
		}
	}
	return mapping
}

func parsePort(rawPort string) (int, error) {
	port, err := strconv.ParseUint(rawPort, 10, 16)
	if err != nil {
		return 0, err
	}
	return int(port), nil
}

// RenameContainer updates and existing containers name
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#rename-a-container for more details.
func (c *Client) RenameContainer(opts RenameContainerOptions) error {
	_, _, err := c.do("POST", fmt.Sprintf("/containers/"+opts.ID+"/rename?%s", queryString(opts)), doOptions{})
	return err
}

// InspectContainer returns information about a container by its ID.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#inspect-a-container for more details.
func (c *Client) InspectContainer(id string) (*Container, error) {
	path := "/containers/" + id + "/json"
	body, status, err := c.do("GET", path, doOptions{})
	if status == http.StatusNotFound {
		return nil, &NoSuchContainer{ID: id}
	}
	if err != nil {
		return nil, err
	}
	var container Container
	err = json.Unmarshal(body, &container)
	if err != nil {
		return nil, err
	}
	return &container, nil
}

// ContainerChanges returns changes in the filesystem of the given container.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#inspect-changes-on-a-container-s-filesystem for more details.
func (c *Client) ContainerChanges(id string) ([]Change, error) {
	path := "/containers/" + id + "/changes"
	body, status, err := c.do("GET", path, doOptions{})
	if status == http.StatusNotFound {
		return nil, &NoSuchContainer{ID: id}
	}
	if err != nil {
		return nil, err
	}
	var changes []Change
	err = json.Unmarshal(body, &changes)
	if err != nil {
		return nil, err
	}
	return changes, nil
}

// CreateContainer creates a new container, returning the container instance,
// or an error in case of failure.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#create-a-container for more details.
func (c *Client) CreateContainer(opts CreateContainerOptions) (*Container, error) {
	path := "/containers/create?" + queryString(opts)
	body, status, err := c.do(
		"POST",
		path,
		doOptions{
			data: struct {
				*Config
				HostConfig *HostConfig `json:"HostConfig,omitempty" yaml:"HostConfig,omitempty"`
			}{
				opts.Config,
				opts.HostConfig,
			},
		},
	)

	if status == http.StatusNotFound {
		return nil, ErrNoSuchImage
	}
	if status == http.StatusConflict {
		return nil, ErrContainerAlreadyExists
	}
	if err != nil {
		return nil, err
	}
	var container Container
	err = json.Unmarshal(body, &container)
	if err != nil {
		return nil, err
	}

	container.Name = opts.Name

	return &container, nil
}

// AlwaysRestart returns a restart policy that tells the Docker daemon to
// always restart the container.
func AlwaysRestart() RestartPolicy {
	return RestartPolicy{Name: "always"}
}

// RestartOnFailure returns a restart policy that tells the Docker daemon to
// restart the container on failures, trying at most maxRetry times.
func RestartOnFailure(maxRetry int) RestartPolicy {
	return RestartPolicy{Name: "on-failure", MaximumRetryCount: maxRetry}
}

// NeverRestart returns a restart policy that tells the Docker daemon to never
// restart the container on failures.
func NeverRestart() RestartPolicy {
	return RestartPolicy{Name: "no"}
}

// StartContainer starts a container, returning an error in case of failure.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#start-a-container for more details.
func (c *Client) StartContainer(id string, hostConfig *HostConfig) error {
	path := "/containers/" + id + "/start"
	_, status, err := c.do("POST", path, doOptions{data: hostConfig, forceJSON: true})
	if status == http.StatusNotFound {
		return &NoSuchContainer{ID: id, Err: err}
	}
	if status == http.StatusNotModified {
		return &ContainerAlreadyRunning{ID: id}
	}
	if err != nil {
		return err
	}
	return nil
}

// StopContainer stops a container, killing it after the given timeout (in
// seconds).
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#stop-a-container for more details.
func (c *Client) StopContainer(id string, timeout uint) error {
	path := fmt.Sprintf("/containers/%s/stop?t=%d", id, timeout)
	_, status, err := c.do("POST", path, doOptions{})
	if status == http.StatusNotFound {
		return &NoSuchContainer{ID: id}
	}
	if status == http.StatusNotModified {
		return &ContainerNotRunning{ID: id}
	}
	if err != nil {
		return err
	}
	return nil
}

// RestartContainer stops a container, killing it after the given timeout (in
// seconds), during the stop process.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#restart-a-container for more details.
func (c *Client) RestartContainer(id string, timeout uint) error {
	path := fmt.Sprintf("/containers/%s/restart?t=%d", id, timeout)
	_, status, err := c.do("POST", path, doOptions{})
	if status == http.StatusNotFound {
		return &NoSuchContainer{ID: id}
	}
	if err != nil {
		return err
	}
	return nil
}

// PauseContainer pauses the given container.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#pause-a-container for more details.
func (c *Client) PauseContainer(id string) error {
	path := fmt.Sprintf("/containers/%s/pause", id)
	_, status, err := c.do("POST", path, doOptions{})
	if status == http.StatusNotFound {
		return &NoSuchContainer{ID: id}
	}
	if err != nil {
		return err
	}
	return nil
}

// UnpauseContainer unpauses the given container.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#unpause-a-container for more details.
func (c *Client) UnpauseContainer(id string) error {
	path := fmt.Sprintf("/containers/%s/unpause", id)
	_, status, err := c.do("POST", path, doOptions{})
	if status == http.StatusNotFound {
		return &NoSuchContainer{ID: id}
	}
	if err != nil {
		return err
	}
	return nil
}

// TopContainer returns processes running inside a container
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#list-processes-running-inside-a-container for more details.
func (c *Client) TopContainer(id string, psArgs string) (TopResult, error) {
	var args string
	var result TopResult
	if psArgs != "" {
		args = fmt.Sprintf("?ps_args=%s", psArgs)
	}
	path := fmt.Sprintf("/containers/%s/top%s", id, args)
	body, status, err := c.do("GET", path, doOptions{})
	if status == http.StatusNotFound {
		return result, &NoSuchContainer{ID: id}
	}
	if err != nil {
		return result, err
	}
	err = json.Unmarshal(body, &result)
	if err != nil {
		return result, err
	}
	return result, nil
}

// Stats sends container statistics for the given container to the given channel.
//
// This function is blocking, similar to a streaming call for logs, and should be run
// on a separate goroutine from the caller. Note that this function will block until
// the given container is removed, not just exited. When finished, this function
// will close the given channel.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#get-container-stats-based-on-resource-usage for more details.
func (c *Client) Stats(opts StatsOptions) (retErr error) {
	errC := make(chan error, 1)
	readCloser, writeCloser := io.Pipe()

	defer func() {
		close(opts.Stats)
		if err := <-errC; err != nil && retErr == nil {
			retErr = err
		}
		if err := readCloser.Close(); err != nil && retErr == nil {
			retErr = err
		}
	}()

	go func() {
		err := c.stream("GET", fmt.Sprintf("/containers/%s/stats?stream=%v", opts.ID, opts.Stream), streamOptions{
			rawJSONStream:  true,
			useJSONDecoder: true,
			stdout:         writeCloser,
		})
		if err != nil {
			dockerError, ok := err.(*Error)
			if ok {
				if dockerError.Status == http.StatusNotFound {
					err = &NoSuchContainer{ID: opts.ID}
				}
			}
		}
		if closeErr := writeCloser.Close(); closeErr != nil && err == nil {
			err = closeErr
		}
		errC <- err
		close(errC)
	}()

	decoder := json.NewDecoder(readCloser)
	stats := new(Stats)
	for err := decoder.Decode(&stats); err != io.EOF; err = decoder.Decode(stats) {
		if err != nil {
			return err
		}
		opts.Stats <- stats
		stats = new(Stats)
	}
	return nil
}

// KillContainer kills a container, returning an error in case of failure.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#kill-a-container for more details.
func (c *Client) KillContainer(opts KillContainerOptions) error {
	path := "/containers/" + opts.ID + "/kill" + "?" + queryString(opts)
	_, status, err := c.do("POST", path, doOptions{})
	if status == http.StatusNotFound {
		return &NoSuchContainer{ID: opts.ID}
	}
	if err != nil {
		return err
	}
	return nil
}

// RemoveContainer removes a container, returning an error in case of failure.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#remove-a-container for more details.
func (c *Client) RemoveContainer(opts RemoveContainerOptions) error {
	path := "/containers/" + opts.ID + "?" + queryString(opts)
	_, status, err := c.do("DELETE", path, doOptions{})
	if status == http.StatusNotFound {
		return &NoSuchContainer{ID: opts.ID}
	}
	if err != nil {
		return err
	}
	return nil
}

// CopyFromContainer copy files or folders from a container, using a given
// resource.
//
// See https://docs.docker.com/reference/api/docker_remote_api_v1.19/#copy-files-or-folders-from-a-container for more details.
func (c *Client) CopyFromContainer(opts CopyFromContainerOptions) error {
	if opts.Container == "" {
		return &NoSuchContainer{ID: opts.Container}
	}
	url := fmt.Sprintf("/containers/%s/copy", opts.Container)
	body, status, err := c.do("POST", url, doOptions{data: opts})
	if status == http.StatusNotFound {
		return &NoSuchContainer{ID: opts.Container}
	}
	if err != nil {
		return err
	}
	_, err = io.Copy(opts.OutputStream, bytes.NewBuffer(body))
	return err
}

// WaitContainer blocks until the given container stops, return the exit code
// of the container status.
//
// See http://goo.gl/J88DHU for more details.
func (c *Client) WaitContainer(id string) (int, error) {
	body, status, err := c.do("POST", "/containers/"+id+"/wait", doOptions{})
	if status == http.StatusNotFound {
		return 0, &NoSuchContainer{ID: id}
	}
	if err != nil {
		return 0, err
	}
	var r struct{ StatusCode int }
	err = json.Unmarshal(body, &r)
	if err != nil {
		return 0, err
	}
	return r.StatusCode, nil
}

// CommitContainer creates a new image from a container's changes.
//
// See http://goo.gl/Jn8pe8 for more details.
func (c *Client) CommitContainer(opts CommitContainerOptions) (*Image, error) {
	path := "/commit?" + queryString(opts)
	body, status, err := c.do("POST", path, doOptions{data: opts.Run})
	if status == http.StatusNotFound {
		return nil, &NoSuchContainer{ID: opts.Container}
	}
	if err != nil {
		return nil, err
	}
	var image Image
	err = json.Unmarshal(body, &image)
	if err != nil {
		return nil, err
	}
	return &image, nil
}

// AttachToContainer attaches to a container, using the given options.
//
// See http://goo.gl/RRAhws for more details.
func (c *Client) AttachToContainer(opts AttachToContainerOptions) error {
	if opts.Container == "" {
		return &NoSuchContainer{ID: opts.Container}
	}
	path := "/containers/" + opts.Container + "/attach?" + queryString(opts)
	return c.hijack("POST", path, hijackOptions{
		success:        opts.Success,
		setRawTerminal: opts.RawTerminal,
		in:             opts.InputStream,
		stdout:         opts.OutputStream,
		stderr:         opts.ErrorStream,
	})
}

// Logs gets stdout and stderr logs from the specified container.
//
// See http://goo.gl/rLhKSU for more details.
func (c *Client) Logs(opts LogsOptions) error {
	if opts.Container == "" {
		return &NoSuchContainer{ID: opts.Container}
	}
	if opts.Tail == "" {
		opts.Tail = "all"
	}
	path := "/containers/" + opts.Container + "/logs?" + queryString(opts)
	return c.stream("GET", path, streamOptions{
		setRawTerminal: opts.RawTerminal,
		stdout:         opts.OutputStream,
		stderr:         opts.ErrorStream,
	})
}

// ResizeContainerTTY resizes the terminal to the given height and width.
func (c *Client) ResizeContainerTTY(id string, height, width int) error {
	params := make(url.Values)
	params.Set("h", strconv.Itoa(height))
	params.Set("w", strconv.Itoa(width))
	_, _, err := c.do("POST", "/containers/"+id+"/resize?"+params.Encode(), doOptions{})
	return err
}

// ExportContainer export the contents of container id as tar archive
// and prints the exported contents to stdout.
//
// See http://goo.gl/hnzE62 for more details.
func (c *Client) ExportContainer(opts ExportContainerOptions) error {
	if opts.ID == "" {
		return &NoSuchContainer{ID: opts.ID}
	}
	url := fmt.Sprintf("/containers/%s/export", opts.ID)
	return c.stream("GET", url, streamOptions{
		setRawTerminal: true,
		stdout:         opts.OutputStream,
	})
}

func (err *NoSuchContainer) Error() string {
	if err.Err != nil {
		return err.Err.Error()
	}
	return "No such container: " + err.ID
}

func (err *ContainerAlreadyRunning) Error() string {
	return "Container already running: " + err.ID
}
func (err *ContainerNotRunning) Error() string {
	return "Container not running: " + err.ID
}

//****************************************************************//
//env need func
//****************************************************************//

// Get returns the string value of the given key.
func (env *Env) Get(key string) (value string) {
	return env.Map()[key]
}

// Exists checks whether the given key is defined in the internal Env
// representation.
func (env *Env) Exists(key string) bool {
	_, exists := env.Map()[key]
	return exists
}

// GetBool returns a boolean representation of the given key. The key is false
// whenever its value if 0, no, false, none or an empty string. Any other value
// will be interpreted as true.
func (env *Env) GetBool(key string) (value bool) {
	s := strings.ToLower(strings.Trim(env.Get(key), " \t"))
	if s == "" || s == "0" || s == "no" || s == "false" || s == "none" {
		return false
	}
	return true
}

// SetBool defines a boolean value to the given key.
func (env *Env) SetBool(key string, value bool) {
	if value {
		env.Set(key, "1")
	} else {
		env.Set(key, "0")
	}
}

// GetInt returns the value of the provided key, converted to int.
//
// It the value cannot be represented as an integer, it returns -1.
func (env *Env) GetInt(key string) int {
	return int(env.GetInt64(key))
}

// SetInt defines an integer value to the given key.
func (env *Env) SetInt(key string, value int) {
	env.Set(key, strconv.Itoa(value))
}

// GetInt64 returns the value of the provided key, converted to int64.
//
// It the value cannot be represented as an integer, it returns -1.
func (env *Env) GetInt64(key string) int64 {
	s := strings.Trim(env.Get(key), " \t")
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return -1
	}
	return val
}

// SetInt64 defines an integer (64-bit wide) value to the given key.
func (env *Env) SetInt64(key string, value int64) {
	env.Set(key, strconv.FormatInt(value, 10))
}

// GetJSON unmarshals the value of the provided key in the provided iface.
//
// iface is a value that can be provided to the json.Unmarshal function.
func (env *Env) GetJSON(key string, iface interface{}) error {
	sval := env.Get(key)
	if sval == "" {
		return nil
	}
	return json.Unmarshal([]byte(sval), iface)
}

// SetJSON marshals the given value to JSON format and stores it using the
// provided key.
func (env *Env) SetJSON(key string, value interface{}) error {
	sval, err := json.Marshal(value)
	if err != nil {
		return err
	}
	env.Set(key, string(sval))
	return nil
}

// GetList returns a list of strings matching the provided key. It handles the
// list as a JSON representation of a list of strings.
//
// If the given key matches to a single string, it will return a list
// containing only the value that matches the key.
func (env *Env) GetList(key string) []string {
	sval := env.Get(key)
	if sval == "" {
		return nil
	}
	var l []string
	if err := json.Unmarshal([]byte(sval), &l); err != nil {
		l = append(l, sval)
	}
	return l
}

// SetList stores the given list in the provided key, after serializing it to
// JSON format.
func (env *Env) SetList(key string, value []string) error {
	return env.SetJSON(key, value)
}

// Set defines the value of a key to the given string.
func (env *Env) Set(key, value string) {
	*env = append(*env, key+"="+value)
}

// Decode decodes `src` as a json dictionary, and adds each decoded key-value
// pair to the environment.
//
// If `src` cannot be decoded as a json dictionary, an error is returned.
func (env *Env) Decode(src io.Reader) error {
	m := make(map[string]interface{})
	if err := json.NewDecoder(src).Decode(&m); err != nil {
		return err
	}
	for k, v := range m {
		env.SetAuto(k, v)
	}
	return nil
}

// SetAuto will try to define the Set* method to call based on the given value.
func (env *Env) SetAuto(key string, value interface{}) {
	if fval, ok := value.(float64); ok {
		env.SetInt64(key, int64(fval))
	} else if sval, ok := value.(string); ok {
		env.Set(key, sval)
	} else if val, err := json.Marshal(value); err == nil {
		env.Set(key, string(val))
	} else {
		env.Set(key, fmt.Sprintf("%v", value))
	}
}

// Map returns the map representation of the env.
func (env *Env) Map() map[string]string {
	if len(*env) == 0 {
		return nil
	}
	m := make(map[string]string)
	for _, kv := range *env {
		parts := strings.SplitN(kv, "=", 2)
		m[parts[0]] = parts[1]
	}
	return m
}

//****************************************************************//
//event need func
//****************************************************************//

// AddEventListener adds a new listener to container events in the Docker API.
//
// The parameter is a channel through which events will be sent.
func (c *Client) AddEventListener(listener chan<- *APIEvents) error {
	var err error
	if !c.eventMonitor.isEnabled() {
		err = c.eventMonitor.enableEventMonitoring(c)
		if err != nil {
			return err
		}
	}
	err = c.eventMonitor.addListener(listener)
	if err != nil {
		return err
	}
	return nil
}

// RemoveEventListener removes a listener from the monitor.
func (c *Client) RemoveEventListener(listener chan *APIEvents) error {
	err := c.eventMonitor.removeListener(listener)
	if err != nil {
		return err
	}
	if len(c.eventMonitor.listeners) == 0 {
		err = c.eventMonitor.disableEventMonitoring()
		if err != nil {
			return err
		}
	}
	return nil
}

func (eventState *eventMonitoringState) addListener(listener chan<- *APIEvents) error {
	eventState.Lock()
	defer eventState.Unlock()
	if listenerExists(listener, &eventState.listeners) {
		return ErrListenerAlreadyExists
	}
	eventState.Add(1)
	eventState.listeners = append(eventState.listeners, listener)
	return nil
}

func (eventState *eventMonitoringState) removeListener(listener chan<- *APIEvents) error {
	eventState.Lock()
	defer eventState.Unlock()
	if listenerExists(listener, &eventState.listeners) {
		var newListeners []chan<- *APIEvents
		for _, l := range eventState.listeners {
			if l != listener {
				newListeners = append(newListeners, l)
			}
		}
		eventState.listeners = newListeners
		eventState.Add(-1)
	}
	return nil
}

func (eventState *eventMonitoringState) closeListeners() {
	eventState.Lock()
	defer eventState.Unlock()
	for _, l := range eventState.listeners {
		close(l)
		eventState.Add(-1)
	}
	eventState.listeners = nil
}

func listenerExists(a chan<- *APIEvents, list *[]chan<- *APIEvents) bool {
	for _, b := range *list {
		if b == a {
			return true
		}
	}
	return false
}

func (eventState *eventMonitoringState) enableEventMonitoring(c *Client) error {
	eventState.Lock()
	defer eventState.Unlock()
	if !eventState.enabled {
		eventState.enabled = true
		var lastSeenDefault = int64(0)
		eventState.lastSeen = &lastSeenDefault
		eventState.C = make(chan *APIEvents, 100)
		eventState.errC = make(chan error, 1)
		go eventState.monitorEvents(c)
	}
	return nil
}

func (eventState *eventMonitoringState) disableEventMonitoring() error {
	eventState.Wait()
	eventState.Lock()
	defer eventState.Unlock()
	if eventState.enabled {
		eventState.enabled = false
		close(eventState.C)
		close(eventState.errC)
	}
	return nil
}

func (eventState *eventMonitoringState) monitorEvents(c *Client) {
	var err error
	for eventState.noListeners() {
		time.Sleep(10 * time.Millisecond)
	}
	if err = eventState.connectWithRetry(c); err != nil {
		eventState.terminate()
	}
	for eventState.isEnabled() {
		timeout := time.After(100 * time.Millisecond)
		select {
		case ev, ok := <-eventState.C:
			if !ok {
				return
			}
			if ev == EOFEvent {
				eventState.closeListeners()
				eventState.terminate()
				return
			}
			eventState.updateLastSeen(ev)
			go eventState.sendEvent(ev)
		case err = <-eventState.errC:
			if err == ErrNoListeners {
				eventState.terminate()
				return
			} else if err != nil {
				defer func() { go eventState.monitorEvents(c) }()
				return
			}
		case <-timeout:
			continue
		}
	}
}

func (eventState *eventMonitoringState) connectWithRetry(c *Client) error {
	var retries int
	var err error
	for err = c.eventHijack(atomic.LoadInt64(eventState.lastSeen), eventState.C, eventState.errC); err != nil && retries < maxMonitorConnRetries; retries++ {
		waitTime := int64(retryInitialWaitTime * math.Pow(2, float64(retries)))
		time.Sleep(time.Duration(waitTime) * time.Millisecond)
		err = c.eventHijack(atomic.LoadInt64(eventState.lastSeen), eventState.C, eventState.errC)
	}
	return err
}

func (eventState *eventMonitoringState) noListeners() bool {
	eventState.RLock()
	defer eventState.RUnlock()
	return len(eventState.listeners) == 0
}

func (eventState *eventMonitoringState) isEnabled() bool {
	eventState.RLock()
	defer eventState.RUnlock()
	return eventState.enabled
}

func (eventState *eventMonitoringState) sendEvent(event *APIEvents) {
	eventState.RLock()
	defer eventState.RUnlock()
	eventState.Add(1)
	defer eventState.Done()
	if eventState.enabled {
		if len(eventState.listeners) == 0 {
			eventState.errC <- ErrNoListeners
			return
		}

		for _, listener := range eventState.listeners {
			listener <- event
		}
	}
}

func (eventState *eventMonitoringState) updateLastSeen(e *APIEvents) {
	eventState.Lock()
	defer eventState.Unlock()
	if atomic.LoadInt64(eventState.lastSeen) < e.Time {
		atomic.StoreInt64(eventState.lastSeen, e.Time)
	}
}

func (eventState *eventMonitoringState) terminate() {
	eventState.disableEventMonitoring()
}

func (c *Client) eventHijack(startTime int64, eventChan chan *APIEvents, errChan chan error) error {
	uri := "/events"
	if startTime != 0 {
		uri += fmt.Sprintf("?since=%d", startTime)
	}
	protocol := c.endpointURL.Scheme
	address := c.endpointURL.Path
	if protocol != "unix" {
		protocol = "tcp"
		address = c.endpointURL.Host
	}
	var dial net.Conn
	var err error
	if c.TLSConfig == nil {
		dial, err = net.Dial(protocol, address)
	} else {
		dial, err = tls.Dial(protocol, address, c.TLSConfig)
	}
	if err != nil {
		return err
	}
	conn := httputil.NewClientConn(dial, nil)
	req, err := http.NewRequest("GET", uri, nil)
	if err != nil {
		return err
	}
	res, err := conn.Do(req)
	if err != nil {
		return err
	}
	go func(res *http.Response, conn *httputil.ClientConn) {
		defer conn.Close()
		defer res.Body.Close()
		decoder := json.NewDecoder(res.Body)
		for {
			var event APIEvents
			if err = decoder.Decode(&event); err != nil {
				if err == io.EOF || err == io.ErrUnexpectedEOF {
					if c.eventMonitor.isEnabled() {
						// Signal that we're exiting.
						eventChan <- EOFEvent
					}
					break
				}
				errChan <- err
			}
			if event.Time == 0 {
				continue
			}
			if !c.eventMonitor.isEnabled() {
				return
			}
			eventChan <- &event
		}
	}(res, conn)
	return nil
}

//****************************************************************//
//exec need func
//****************************************************************//

// CreateExec sets up an exec instance in a running container `id`, returning the exec
// instance, or an error in case of failure.
//
// See http://goo.gl/8izrzI for more details
func (c *Client) CreateExec(opts CreateExecOptions) (*Exec, error) {
	path := fmt.Sprintf("/containers/%s/exec", opts.Container)
	body, status, err := c.do("POST", path, doOptions{data: opts})
	if status == http.StatusNotFound {
		return nil, &NoSuchContainer{ID: opts.Container}
	}
	if err != nil {
		return nil, err
	}
	var exec Exec
	err = json.Unmarshal(body, &exec)
	if err != nil {
		return nil, err
	}

	return &exec, nil
}

// StartExec starts a previously set up exec instance id. If opts.Detach is
// true, it returns after starting the exec command. Otherwise, it sets up an
// interactive session with the exec command.
//
// See http://goo.gl/JW8Lxl for more details
func (c *Client) StartExec(id string, opts StartExecOptions) error {
	if id == "" {
		return &NoSuchExec{ID: id}
	}

	path := fmt.Sprintf("/exec/%s/start", id)

	if opts.Detach {
		_, status, err := c.do("POST", path, doOptions{data: opts})
		if status == http.StatusNotFound {
			return &NoSuchExec{ID: id}
		}
		if err != nil {
			return err
		}
		return nil
	}

	return c.hijack("POST", path, hijackOptions{
		success:        opts.Success,
		setRawTerminal: opts.RawTerminal,
		in:             opts.InputStream,
		stdout:         opts.OutputStream,
		stderr:         opts.ErrorStream,
		data:           opts,
	})
}

// ResizeExecTTY resizes the tty session used by the exec command id. This API
// is valid only if Tty was specified as part of creating and starting the exec
// command.
//
// See http://goo.gl/YDSx1f for more details
func (c *Client) ResizeExecTTY(id string, height, width int) error {
	params := make(url.Values)
	params.Set("h", strconv.Itoa(height))
	params.Set("w", strconv.Itoa(width))

	path := fmt.Sprintf("/exec/%s/resize?%s", id, params.Encode())
	_, _, err := c.do("POST", path, doOptions{})
	return err
}

// InspectExec returns low-level information about the exec command id.
//
// See http://goo.gl/ypQULN for more details
func (c *Client) InspectExec(id string) (*ExecInspect, error) {
	path := fmt.Sprintf("/exec/%s/json", id)
	body, status, err := c.do("GET", path, doOptions{})
	if status == http.StatusNotFound {
		return nil, &NoSuchExec{ID: id}
	}
	if err != nil {
		return nil, err
	}
	var exec ExecInspect
	err = json.Unmarshal(body, &exec)
	if err != nil {
		return nil, err
	}
	return &exec, nil
}

func (err *NoSuchExec) Error() string {
	return "No such exec instance: " + err.ID
}

//****************************************************************//
//image need func
//****************************************************************//

// ListImages returns the list of available images in the server.
//
// See http://goo.gl/HRVN1Z for more details.
func (c *Client) ListImages(opts ListImagesOptions) ([]APIImages, error) {
	path := "/images/json?" + queryString(opts)
	body, _, err := c.do("GET", path, doOptions{})
	if err != nil {
		return nil, err
	}
	var images []APIImages
	err = json.Unmarshal(body, &images)
	if err != nil {
		return nil, err
	}
	return images, nil
}

// ImageHistory returns the history of the image by its name or ID.
//
// See http://goo.gl/2oJmNs for more details.
func (c *Client) ImageHistory(name string) ([]ImageHistory, error) {
	body, status, err := c.do("GET", "/images/"+name+"/history", doOptions{})
	if status == http.StatusNotFound {
		return nil, ErrNoSuchImage
	}
	if err != nil {
		return nil, err
	}
	var history []ImageHistory
	err = json.Unmarshal(body, &history)
	if err != nil {
		return nil, err
	}
	return history, nil
}

// RemoveImage removes an image by its name or ID.
//
// See http://goo.gl/znj0wM for more details.
func (c *Client) RemoveImage(name string) error {
	_, status, err := c.do("DELETE", "/images/"+name, doOptions{})
	if status == http.StatusNotFound {
		return ErrNoSuchImage
	}
	return err
}

// RemoveImageExtended removes an image by its name or ID.
// Extra params can be passed, see RemoveImageOptions
//
// See http://goo.gl/znj0wM for more details.
func (c *Client) RemoveImageExtended(name string, opts RemoveImageOptions) error {
	uri := fmt.Sprintf("/images/%s?%s", name, queryString(&opts))
	_, status, err := c.do("DELETE", uri, doOptions{})
	if status == http.StatusNotFound {
		return ErrNoSuchImage
	}
	return err
}

// InspectImage returns an image by its name or ID.
//
// See http://goo.gl/Q112NY for more details.
func (c *Client) InspectImage(name string) (*Image, error) {
	body, status, err := c.do("GET", "/images/"+name+"/json", doOptions{})
	if status == http.StatusNotFound {
		return nil, ErrNoSuchImage
	}
	if err != nil {
		return nil, err
	}

	var image Image

	// if the caller elected to skip checking the server's version, assume it's the latest
	if c.SkipServerVersionCheck || c.expectedAPIVersion.GreaterThanOrEqualTo(apiVersion112) {
		err = json.Unmarshal(body, &image)
		if err != nil {
			return nil, err
		}
	} else {
		var imagePre012 ImagePre012
		err = json.Unmarshal(body, &imagePre012)
		if err != nil {
			return nil, err
		}

		image.ID = imagePre012.ID
		image.Parent = imagePre012.Parent
		image.Comment = imagePre012.Comment
		image.Created = imagePre012.Created
		image.Container = imagePre012.Container
		image.ContainerConfig = imagePre012.ContainerConfig
		image.DockerVersion = imagePre012.DockerVersion
		image.Author = imagePre012.Author
		image.Config = imagePre012.Config
		image.Architecture = imagePre012.Architecture
		image.Size = imagePre012.Size
	}

	return &image, nil
}

// PushImage pushes an image to a remote registry, logging progress to w.
//
// An empty instance of AuthConfiguration may be used for unauthenticated
// pushes.
//
// See http://goo.gl/pN8A3P for more details.
func (c *Client) PushImage(opts PushImageOptions, auth AuthConfiguration) error {
	if opts.Name == "" {
		return ErrNoSuchImage
	}
	headers, err := headersWithAuth(auth)
	if err != nil {
		return err
	}
	name := opts.Name
	opts.Name = ""
	path := "/images/" + name + "/push?" + queryString(&opts)
	return c.stream("POST", path, streamOptions{
		setRawTerminal: true,
		rawJSONStream:  opts.RawJSONStream,
		headers:        headers,
		stdout:         opts.OutputStream,
	})
}

// PullImage pulls an image from a remote registry, logging progress to opts.OutputStream.
//
// See http://goo.gl/ACyYNS for more details.
func (c *Client) PullImage(opts PullImageOptions, auth AuthConfiguration) error {
	if opts.Repository == "" {
		return ErrNoSuchImage
	}

	headers, err := headersWithAuth(auth)
	if err != nil {
		return err
	}
	return c.createImage(queryString(&opts), headers, nil, opts.OutputStream, opts.RawJSONStream)
}

func (c *Client) createImage(qs string, headers map[string]string, in io.Reader, w io.Writer, rawJSONStream bool) error {
	path := "/images/create?" + qs
	return c.stream("POST", path, streamOptions{
		setRawTerminal: true,
		rawJSONStream:  rawJSONStream,
		headers:        headers,
		in:             in,
		stdout:         w,
	})
}

// LoadImage imports a tarball docker image
//
// See http://goo.gl/Y8NNCq for more details.
func (c *Client) LoadImage(opts LoadImageOptions) error {
	return c.stream("POST", "/images/load", streamOptions{
		setRawTerminal: true,
		in:             opts.InputStream,
	})
}

// ExportImage exports an image (as a tar file) into the stream
//
// See http://goo.gl/mi6kvk for more details.
func (c *Client) ExportImage(opts ExportImageOptions) error {
	return c.stream("GET", fmt.Sprintf("/images/%s/get", opts.Name), streamOptions{
		setRawTerminal: true,
		stdout:         opts.OutputStream,
	})
}

// ExportImages exports one or more images (as a tar file) into the stream
//
// See http://goo.gl/YeZzQK for more details.
func (c *Client) ExportImages(opts ExportImagesOptions) error {
	if opts.Names == nil || len(opts.Names) == 0 {
		return ErrMustSpecifyNames
	}
	return c.stream("GET", "/images/get?"+queryString(&opts), streamOptions{
		setRawTerminal: true,
		stdout:         opts.OutputStream,
	})
}

// ImportImage imports an image from a url, a file or stdin
//
// See http://goo.gl/PhBKnS for more details.
func (c *Client) ImportImage(opts ImportImageOptions) error {
	if opts.Repository == "" {
		return ErrNoSuchImage
	}
	if opts.Source != "-" {
		opts.InputStream = nil
	}
	if opts.Source != "-" && !isURL(opts.Source) {
		f, err := os.Open(opts.Source)
		if err != nil {
			return err
		}
		b, err := ioutil.ReadAll(f)
		opts.InputStream = bytes.NewBuffer(b)
		opts.Source = "-"
	}
	return c.createImage(queryString(&opts), nil, opts.InputStream, opts.OutputStream, opts.RawJSONStream)
}

// BuildImage builds an image from a tarball's url or a Dockerfile in the input
// stream.
//
// See http://goo.gl/7nuGXa for more details.
func (c *Client) BuildImage(opts BuildImageOptions) error {
	if opts.OutputStream == nil {
		return ErrMissingOutputStream
	}
	headers, err := headersWithAuth(opts.Auth, opts.AuthConfigs)
	if err != nil {
		return err
	}

	if opts.Remote != "" && opts.Name == "" {
		opts.Name = opts.Remote
	}
	if opts.InputStream != nil || opts.ContextDir != "" {
		headers["Content-Type"] = "application/tar"
	} else if opts.Remote == "" {
		return ErrMissingRepo
	}
	if opts.ContextDir != "" {
		if opts.InputStream != nil {
			return ErrMultipleContexts
		}
		var err error
		if opts.InputStream, err = createTarStream(opts.ContextDir, opts.Dockerfile); err != nil {
			return err
		}
	}

	return c.stream("POST", fmt.Sprintf("/build?%s", queryString(&opts)), streamOptions{
		setRawTerminal: true,
		rawJSONStream:  opts.RawJSONStream,
		headers:        headers,
		in:             opts.InputStream,
		stdout:         opts.OutputStream,
	})
}

// TagImage adds a tag to the image identified by the given name.
//
// See http://goo.gl/5g6qFy for more details.
func (c *Client) TagImage(name string, opts TagImageOptions) error {
	if name == "" {
		return ErrNoSuchImage
	}
	_, status, err := c.do("POST", fmt.Sprintf("/images/"+name+"/tag?%s",
		queryString(&opts)), doOptions{})

	if status == http.StatusNotFound {
		return ErrNoSuchImage
	}

	return err
}

func isURL(u string) bool {
	p, err := url.Parse(u)
	if err != nil {
		return false
	}
	return p.Scheme == "http" || p.Scheme == "https"
}

func headersWithAuth(auths ...interface{}) (map[string]string, error) {
	var headers = make(map[string]string)

	for _, auth := range auths {
		switch auth.(type) {
		case AuthConfiguration:
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(auth); err != nil {
				return nil, err
			}
			headers["X-Registry-Auth"] = base64.URLEncoding.EncodeToString(buf.Bytes())
		case AuthConfigurations:
			var buf bytes.Buffer
			if err := json.NewEncoder(&buf).Encode(auth); err != nil {
				return nil, err
			}
			headers["X-Registry-Config"] = base64.URLEncoding.EncodeToString(buf.Bytes())
		}
	}

	return headers, nil
}

// SearchImages search the docker hub with a specific given term.
//
// See http://goo.gl/xI5lLZ for more details.
func (c *Client) SearchImages(term string) ([]APIImageSearch, error) {
	body, _, err := c.do("GET", "/images/search?term="+term, doOptions{})
	if err != nil {
		return nil, err
	}
	var searchResult []APIImageSearch
	err = json.Unmarshal(body, &searchResult)
	if err != nil {
		return nil, err
	}
	return searchResult, nil
}

//****************************************************************//
//misc need func
//****************************************************************//

// Version returns version information about the docker server.
//
// See http://goo.gl/BOZrF5 for more details.
func (c *Client) Version() (*Env, error) {
	body, _, err := c.do("GET", "/version", doOptions{})
	if err != nil {
		return nil, err
	}
	var env Env
	if err := env.Decode(bytes.NewReader(body)); err != nil {
		return nil, err
	}
	return &env, nil
}

// Info returns system-wide information about the Docker server.
//
// See http://goo.gl/wmqZsW for more details.
func (c *Client) Info() (*Env, error) {
	body, _, err := c.do("GET", "/info", doOptions{})
	if err != nil {
		return nil, err
	}
	var info Env
	err = info.Decode(bytes.NewReader(body))
	if err != nil {
		return nil, err
	}
	return &info, nil
}

// ParseRepositoryTag gets the name of the repository and returns it splitted
// in two parts: the repository and the tag.
//
// Some examples:
//
//     localhost.localdomain:5000/samalba/hipache:latest -> localhost.localdomain:5000/samalba/hipache, latest
//     localhost.localdomain:5000/samalba/hipache -> localhost.localdomain:5000/samalba/hipache, ""
func ParseRepositoryTag(repoTag string) (repository string, tag string) {
	n := strings.LastIndex(repoTag, ":")
	if n < 0 {
		return repoTag, ""
	}
	if tag := repoTag[n+1:]; !strings.Contains(tag, "/") {
		return repoTag[:n], tag
	}
	return repoTag, ""
}

//****************************************************************//
//tar need func
//****************************************************************//

func createTarStream(srcPath, dockerfilePath string) (io.ReadCloser, error) {
	excludes, err := parseDockerignore(srcPath)
	if err != nil {
		return nil, err
	}

	includes := []string{"."}

	// If .dockerignore mentions .dockerignore or the Dockerfile
	// then make sure we send both files over to the daemon
	// because Dockerfile is, obviously, needed no matter what, and
	// .dockerignore is needed to know if either one needs to be
	// removed.  The deamon will remove them for us, if needed, after it
	// parses the Dockerfile.
	//
	// https://github.com/docker/docker/issues/8330
	//
	forceIncludeFiles := []string{".dockerignore", dockerfilePath}

	for _, includeFile := range forceIncludeFiles {
		if includeFile == "" {
			continue
		}
		keepThem, err := fileutils.Matches(includeFile, excludes)
		if err != nil {
			return nil, fmt.Errorf("cannot match .dockerfile: '%s', error: %s", includeFile, err)
		}
		if keepThem {
			includes = append(includes, includeFile)
		}
	}

	if err := validateContextDirectory(srcPath, excludes); err != nil {
		return nil, err
	}
	tarOpts := &archive.TarOptions{
		ExcludePatterns: excludes,
		IncludeFiles:    includes,
		Compression:     archive.Uncompressed,
		NoLchown:        true,
	}
	return archive.TarWithOptions(srcPath, tarOpts)
}

// validateContextDirectory checks if all the contents of the directory
// can be read and returns an error if some files can't be read.
// Symlinks which point to non-existing files don't trigger an error
func validateContextDirectory(srcPath string, excludes []string) error {
	return filepath.Walk(filepath.Join(srcPath, "."), func(filePath string, f os.FileInfo, err error) error {
		// skip this directory/file if it's not in the path, it won't get added to the context
		if relFilePath, err := filepath.Rel(srcPath, filePath); err != nil {
			return err
		} else if skip, err := fileutils.Matches(relFilePath, excludes); err != nil {
			return err
		} else if skip {
			if f.IsDir() {
				return filepath.SkipDir
			}
			return nil
		}

		if err != nil {
			if os.IsPermission(err) {
				return fmt.Errorf("can't stat '%s'", filePath)
			}
			if os.IsNotExist(err) {
				return nil
			}
			return err
		}

		// skip checking if symlinks point to non-existing files, such symlinks can be useful
		// also skip named pipes, because they hanging on open
		if f.Mode()&(os.ModeSymlink|os.ModeNamedPipe) != 0 {
			return nil
		}

		if !f.IsDir() {
			currentFile, err := os.Open(filePath)
			if err != nil && os.IsPermission(err) {
				return fmt.Errorf("no permission to read from '%s'", filePath)
			}
			currentFile.Close()
		}
		return nil
	})
}

func parseDockerignore(root string) ([]string, error) {
	var excludes []string
	ignore, err := ioutil.ReadFile(path.Join(root, ".dockerignore"))
	if err != nil && !os.IsNotExist(err) {
		return excludes, fmt.Errorf("error reading .dockerignore: '%s'", err)
	}
	excludes = strings.Split(string(ignore), "\n")

	return excludes, nil
}

//****************************************************************//
//tls need func
//****************************************************************//

func (c *tlsClientCon) CloseWrite() error {
	// Go standard tls.Conn doesn't provide the CloseWrite() method so we do it
	// on its underlying connection.
	if cwc, ok := c.rawConn.(interface {
		CloseWrite() error
	}); ok {
		return cwc.CloseWrite()
	}
	return nil
}

func tlsDialWithDialer(dialer *net.Dialer, network, addr string, config *tls.Config) (net.Conn, error) {
	// We want the Timeout and Deadline values from dialer to cover the
	// whole process: TCP connection and TLS handshake. This means that we
	// also need to start our own timers now.
	timeout := dialer.Timeout

	if !dialer.Deadline.IsZero() {
		deadlineTimeout := dialer.Deadline.Sub(time.Now())
		if timeout == 0 || deadlineTimeout < timeout {
			timeout = deadlineTimeout
		}
	}

	var errChannel chan error

	if timeout != 0 {
		errChannel = make(chan error, 2)
		time.AfterFunc(timeout, func() {
			errChannel <- errors.New("")
		})
	}

	rawConn, err := dialer.Dial(network, addr)
	if err != nil {
		return nil, err
	}

	colonPos := strings.LastIndex(addr, ":")
	if colonPos == -1 {
		colonPos = len(addr)
	}
	hostname := addr[:colonPos]

	// If no ServerName is set, infer the ServerName
	// from the hostname we're connecting to.
	if config.ServerName == "" {
		// Make a copy to avoid polluting argument or default.
		c := *config
		c.ServerName = hostname
		config = &c
	}

	conn := tls.Client(rawConn, config)

	if timeout == 0 {
		err = conn.Handshake()
	} else {
		go func() {
			errChannel <- conn.Handshake()
		}()

		err = <-errChannel
	}

	if err != nil {
		rawConn.Close()
		return nil, err
	}

	// This is Docker difference with standard's crypto/tls package: returned a
	// wrapper which holds both the TLS and raw connections.
	return &tlsClientCon{conn, rawConn}, nil
}

func tlsDial(network, addr string, config *tls.Config) (net.Conn, error) {
	return tlsDialWithDialer(new(net.Dialer), network, addr, config)
}
