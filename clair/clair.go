package clair

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/coreos/clair/database"
	"github.com/coreos/clair/updater"
	_ "github.com/coreos/clair/updater/fetchers"
	"github.com/coreos/clair/utils"
	"github.com/coreos/clair/utils/types"
	"github.com/coreos/clair/worker"
	_ "github.com/coreos/clair/worker/detectors/os"
	_ "github.com/coreos/clair/worker/detectors/packages"
	"github.com/coreos/pkg/capnslog"

	"github.com/containerops/dockyard/utils/setting"
)

type ClairConfig struct {
	KeepDB       bool
	DBPath       string
	LogLevel     string
	Duration     string
	VulnPriority string
}

type History struct {
	ID        string
	Parent    string
	localPath string
}

const (
	DefaultClairUpdateDuration = "1h0m0s"
	DefaultClairLogLevel       = "info"
	DefaultClairDBPath         = "/db"
	DefaultClairVulnPriority   = "Low"
)

var (
	clairConf    ClairConfig
	clairStopper *utils.Stopper
)

var ClairSc ShareChannel

func InitClair() error {
	ClairSc = *NewShareChannel()

	if setting.ClairDBPath != "" {
		clairConf.DBPath = setting.ClairDBPath
	} else {
		clairConf.DBPath = DefaultClairDBPath
	}

	clairConf.KeepDB = setting.ClairKeepDB
	clairConf.LogLevel = setting.ClairLogLevel
	clairConf.Duration = setting.ClairUpdateDuration
	clairConf.VulnPriority = setting.ClairVulnPriority

	if err := database.Open("bolt", clairConf.DBPath); err != nil {
		logrus.Debug(err)
		return err
	}

	logLevel, err := capnslog.ParseLevel(strings.ToUpper(clairConf.LogLevel))
	if err != nil {
		logLevel, _ = capnslog.ParseLevel(strings.ToUpper(DefaultClairLogLevel))
	}
	capnslog.SetGlobalLogLevel(logLevel)
	capnslog.SetFormatter(capnslog.NewPrettyFormatter(os.Stdout, false))

	if types.Priority(clairConf.VulnPriority).IsValid() {
		logrus.Debugf("Vuln priority is invalid :%v.", clairConf.VulnPriority)
		clairConf.VulnPriority = DefaultClairVulnPriority
	}

	if clairConf.Duration == "" {
		logrus.Debugf("No duration set, so only update at the beginning.")
		go updater.Update()
		clairStopper = nil
	} else {
		st := utils.NewStopper()
		st.Begin()
		d, err := time.ParseDuration(clairConf.Duration)
		if err != nil {
			logrus.Warnf("Wrong duration format, use the default duration: %v.", DefaultClairUpdateDuration)
			clairConf.Duration = DefaultClairUpdateDuration
			d, err = time.ParseDuration(clairConf.Duration)
			if err != nil {
				logrus.Debugf("Cannot pare du %v", err)
			}
		}

		go updater.Run(d, st)
		clairStopper = st
		st.Begin()
	}
	return nil
}

func StopClair() {
	if clairStopper != nil {
		clairStopper.End()
	}
	if !clairConf.KeepDB {
		os.RemoveAll(clairConf.DBPath)
	}

	database.Close()
}

func Put(manifest []byte, vendor string, version string) (err error) {
	fmt.Println("start to scan")
	var history []History
	if vendor == "Docker" && version == "V2" {
		if history, err = getDockerV2History(manifest); err != nil {
			return err
		}
	}

	//The first one should be the base image (no parent)
	for i := 0; i < len(history); i++ {
		PutLayer(history[i])
	}
	return nil
}

func PutLayer(h History) error {
	if err := worker.Process(h.ID, h.Parent, h.localPath); err != nil {
		logrus.Debugf("End find err process: %v", err)
		return err
	}
	return nil
}

func Get(manifest []byte, vendor string, version string) (vulns []*database.Vulnerability, err error) {
	var history []History
	if vendor == "Docker" && version == "V2" {
		if history, err = getDockerV2History(manifest); err != nil {
			return nil, err
		}
	}
	if len(history) > 0 {
		return GetVulns(history[len(history)-1].ID)
	} else {
		return nil, nil
	}
}

func GetVulns(ID string) ([]*database.Vulnerability, error) {
	layer, err := database.FindOneLayerByID(ID, []string{database.FieldLayerParent, database.FieldLayerPackages})
	if err != nil {
		logrus.Debugf("Cannot get layer: %v", err)
		return nil, err
	}

	packagesNodes, err := layer.AllPackages()
	if err != nil {
		logrus.Debugf("Cannot get packages: %v", err)
		return nil, err
	}

	return getVulnerabilitiesFromLayerPackagesNodes(packagesNodes, types.Priority(clairConf.VulnPriority), []string{database.FieldVulnerabilityID, database.FieldVulnerabilityLink, database.FieldVulnerabilityPriority, database.FieldVulnerabilityDescription})
}

func getVulnerabilitiesFromLayerPackagesNodes(packagesNodes []string, minimumPriority types.Priority, selectedFields []string) ([]*database.Vulnerability, error) {
	if len(packagesNodes) == 0 {
		return []*database.Vulnerability{}, nil
	}

	packagesNextVersions, err := getSuccessorsFromPackagesNodes(packagesNodes)
	if err != nil {
		return []*database.Vulnerability{}, err
	}
	if len(packagesNextVersions) == 0 {
		return []*database.Vulnerability{}, nil
	}

	vulnerabilities, err := database.FindAllVulnerabilitiesByFixedIn(packagesNextVersions, selectedFields)
	if err != nil {
		return []*database.Vulnerability{}, err
	}

	filteredVulnerabilities := []*database.Vulnerability{}
	seen := map[string]struct{}{}
	for _, v := range vulnerabilities {
		if minimumPriority.Compare(v.Priority) <= 0 {
			if _, alreadySeen := seen[v.ID]; !alreadySeen {
				filteredVulnerabilities = append(filteredVulnerabilities, v)
				seen[v.ID] = struct{}{}
			}
		}
	}

	return filteredVulnerabilities, nil
}

func getSuccessorsFromPackagesNodes(packagesNodes []string) ([]string, error) {
	if len(packagesNodes) == 0 {
		return []string{}, nil
	}

	packages, err := database.FindAllPackagesByNodes(packagesNodes, []string{database.FieldPackageNextVersion})
	if err != nil {
		return []string{}, err
	}

	var packagesNextVersions []string
	for _, pkg := range packages {
		nextVersions, err := pkg.NextVersions([]string{})
		if err != nil {
			return []string{}, err
		}
		for _, version := range nextVersions {
			packagesNextVersions = append(packagesNextVersions, version.Node)
		}
	}

	return packagesNextVersions, nil
}

//V2 only
func getDockerV2History(data []byte) (history []History, err error) {
	fmt.Println("get history of ", string(data))
	var manifest map[string]interface{}
	if err := json.Unmarshal(data, &manifest); err != nil {
		return nil, err
	}
	//In manifest, the last one is the base image
	for k := len(manifest["history"].([]interface{})) - 1; k >= 0; k-- {
		var h History
		compatibility := manifest["history"].([]interface{})[k].(map[string]interface{})["v1Compatibility"].(string)
		if err := json.Unmarshal([]byte(compatibility), &h); err != nil {
			return nil, err
		}

		if len(history) == 0 || history[len(history)-1].ID == h.Parent {
			digest := manifest["fsLayers"].([]interface{})[k].(map[string]interface{})["blobSum"].(string)
			h.localPath = getDockerV2Path(digest)
			history = append(history, h)
		}
	}

	return history, nil
}

func getDockerV2Path(digest string) string {
	tarsum := strings.Split(digest, ":")[1]
	layerfile := fmt.Sprintf("%v/tarsum/%v/layer", setting.ImagePath, tarsum)
	return layerfile
}
