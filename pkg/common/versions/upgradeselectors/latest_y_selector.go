package upgradeselectors

import (
	"fmt"
	"strings"

	"github.com/openshift/osde2e/pkg/common/config"
	"github.com/openshift/osde2e/pkg/common/spi"
	"github.com/spf13/viper"
)

func init() {
	registerSelector(latestYVersion{})
}

// latestYVersion returns the next minor version upgrade available
type latestYVersion struct{}

func (l latestYVersion) ShouldUse() bool {
	return viper.GetBool(config.Upgrade.UpgradeToLatestY)
}

func (l latestYVersion) Priority() int {
	return 70
}

func (l latestYVersion) SelectVersion(installVersion *spi.Version, versionList *spi.VersionList) (*spi.Version, string, error) {
	var newestVersion *spi.Version
	newestVersion = installVersion

	for _, v := range versionList.FindVersion(installVersion.Version().Original()) {
		for upgradeVersion := range v.AvailableUpgrades() {
			if upgradeVersion.Minor() != installVersion.Version().Minor()+1 {
				continue
			}

			// Automatically assume a Y+1 nightly is greater than a non-nightly-build
			if strings.Contains(upgradeVersion.Original(), "nightly") && !strings.Contains(newestVersion.Version().Original(), "nightly") {
				newestVersion = spi.NewVersionBuilder().Version(upgradeVersion).Build()
				continue
			}

			// Automatically assume a Y+1 nightly is greater than a non-nightly-build
			if strings.Contains(upgradeVersion.Original(), "nightly") && strings.Contains(newestVersion.Version().Original(), "nightly") {
				if upgradeVersion.Original() > newestVersion.Version().Original() {
					newestVersion = spi.NewVersionBuilder().Version(upgradeVersion).Build()
				}
			} else {
				// Catch the rest
				if upgradeVersion.GreaterThan(newestVersion.Version()) {
					newestVersion = spi.NewVersionBuilder().Version(upgradeVersion).Build()
				}
			}
		}
	}

	if !newestVersion.Version().GreaterThan(installVersion.Version()) {
		return nil, "latest y version", fmt.Errorf("No available upgrade path for version %s", installVersion.Version().Original())
	}
	return newestVersion, "latest y version", nil
}
