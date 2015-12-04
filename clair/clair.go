package clair

import (
	"os"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/containerops/wrench/setting"
	"github.com/coreos/clair/database"
	"github.com/coreos/clair/updater"
	"github.com/coreos/clair/utils"
	"github.com/coreos/clair/utils/types"
	"github.com/coreos/clair/worker"
	"github.com/coreos/pkg/capnslog"

	_ "github.com/coreos/clair/updater/fetchers"
	_ "github.com/coreos/clair/worker/detectors/os"
	_ "github.com/coreos/clair/worker/detectors/packages"
)

type ClairConfig struct {
	KeepDB       bool
	DBPath       string
	LogLevel     string
	Duration     string
	VulnPriority string
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

func init() {
	var conf ClairConfig
	// Load database setting
	if setting.ClairDBPath != "" {
		conf.DBPath = setting.ClairDBPath
	} else {
		conf.DBPath = DefaultClairDBPath
	}

	conf.KeepDB = setting.ClairKeepDB
	conf.LogLevel = setting.ClairLogLevel
	conf.Duration = setting.ClairUpdateDuration
	conf.VulnPriority = setting.ClairVulnPriority

	if err := ClairServiceInit(conf); err != nil {
		logrus.Warnf("Cannot init clair service: %v.", err)
	}
}

func ClairServiceInit(conf ClairConfig) error {
	// Set database
	if err := database.Open("bolt", conf.DBPath); err != nil {
		logrus.Debug(err)
		return err
	}

	// Set logLevel of clair lib
	logLevel, err := capnslog.ParseLevel(strings.ToUpper(conf.LogLevel))
	if err != nil {
		logLevel, _ = capnslog.ParseLevel(strings.ToUpper(DefaultClairLogLevel))
	}
	capnslog.SetGlobalLogLevel(logLevel)
	capnslog.SetFormatter(capnslog.NewPrettyFormatter(os.Stdout, false))

	// Set minumum priority parameter.
	if types.Priority(conf.VulnPriority).IsValid() {
		logrus.Debugf("Vuln priority is invalid :%v.", conf.VulnPriority)
		conf.VulnPriority = DefaultClairVulnPriority
	}

	// Set 'duration' and Update the CVE database
	if conf.Duration == "" {
		logrus.Debugf("No duration set, so only update at the beginning.")
		go updater.Update()
		clairStopper = nil
	} else {
		st := utils.NewStopper()
		st.Begin()
		d, err := time.ParseDuration(conf.Duration)
		if err != nil {
			logrus.Warnf("Wrong duration format, use the default duration: %v.", DefaultClairUpdateDuration)
			conf.Duration = DefaultClairUpdateDuration
			d, err = time.ParseDuration(conf.Duration)
			if err != nil {
				logrus.Debugf("Cannot pare du %v", err)
			}
		}

		go updater.Run(d, st)
		clairStopper = st
		st.Begin()
	}
	clairConf = conf
	return nil
}

func ClairServiceStop() {
	if clairStopper != nil {
		clairStopper.End()
	}
	// Remove the database file
	if !clairConf.KeepDB {
		os.RemoveAll(clairConf.DBPath)
	}

	//Bugs in Clair upstream
	//database.Close()
}

func ClairGetVulns(ID string, ParentID string, Path string) ([]*database.Vulnerability, error) {
	// Process data.
	logrus.Debugf("Start to get vulnerabilities: %v %v %v", ID, ParentID, Path)
	if err := worker.Process(ID, ParentID, Path); err != nil {
		logrus.Debugf("End find err process: %v", err)
		return nil, err
	}
	// Find layer
	layer, err := database.FindOneLayerByID(ID, []string{database.FieldLayerParent, database.FieldLayerPackages})
	if err != nil {
		logrus.Debugf("Cannot get layer: %v", err)
		return nil, err
	}

	// Find layer's packages.
	packagesNodes, err := layer.AllPackages()
	if err != nil {
		logrus.Debugf("Cannot get packages: %v", err)
		return nil, err
	}

	// Find vulnerabilities.
	return getVulnerabilitiesFromLayerPackagesNodes(packagesNodes, types.Priority(clairConf.VulnPriority), []string{database.FieldVulnerabilityID, database.FieldVulnerabilityLink, database.FieldVulnerabilityPriority, database.FieldVulnerabilityDescription})
}

// getVulnerabilitiesFromLayerPackagesNodes returns the list of vulnerabilities
// affecting the provided package nodes, filtered by Priority.
func getVulnerabilitiesFromLayerPackagesNodes(packagesNodes []string, minimumPriority types.Priority, selectedFields []string) ([]*database.Vulnerability, error) {
	if len(packagesNodes) == 0 {
		return []*database.Vulnerability{}, nil
	}

	// Get successors of the packages.
	packagesNextVersions, err := getSuccessorsFromPackagesNodes(packagesNodes)
	if err != nil {
		return []*database.Vulnerability{}, err
	}
	if len(packagesNextVersions) == 0 {
		return []*database.Vulnerability{}, nil
	}

	// Find vulnerabilities fixed in these successors.
	vulnerabilities, err := database.FindAllVulnerabilitiesByFixedIn(packagesNextVersions, selectedFields)
	if err != nil {
		return []*database.Vulnerability{}, err
	}

	// Filter vulnerabilities depending on their priority and remove duplicates.
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

// getSuccessorsFromPackagesNodes returns the node list of packages that have
// versions following the versions of the provided packages.
func getSuccessorsFromPackagesNodes(packagesNodes []string) ([]string, error) {
	if len(packagesNodes) == 0 {
		return []string{}, nil
	}

	// Get packages.
	packages, err := database.FindAllPackagesByNodes(packagesNodes, []string{database.FieldPackageNextVersion})
	if err != nil {
		return []string{}, err
	}

	// Find all packages' successors.
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
