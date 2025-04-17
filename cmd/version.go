package cmd

import (
	"fmt"
	"github.com/fatih/color"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/savioxavier/termlink"

	"github.com/BussanQ/kubecm/pkg/update"
	"github.com/cli/safeexec"

	v "github.com/BussanQ/kubecm/version"
	"github.com/spf13/cobra"
)

// VersionCommand version cmd struct
type VersionCommand struct {
	BaseCommand
}

// version returns the version of kubecm.
type version struct {
	// kubecmVersion is a kubecm binary version.
	KubecmVersion string `json:"kubecmVersion"`
	// GoOs holds OS name.
	GoOs string `json:"goOs"`
	// GoArch holds architecture name.
	GoArch string `json:"goArch"`
}

// Init VersionCommand
func (vc *VersionCommand) Init() {
	vc.command = &cobra.Command{
		Use:     "version",
		Short:   "Print version info",
		Long:    "Print version info",
		Aliases: []string{"v"},
		Run: func(cmd *cobra.Command, args []string) {
			kubecmVersion := getVersion().KubecmVersion

			updateMessageChan := make(chan *update.ReleaseInfo)
			go func() {
				rel, _ := update.CheckForUpdate("BussanQ/kubecm", kubecmVersion)
				updateMessageChan <- rel
			}()
			fmt.Printf("%s: %s\n",
				color.BlueString("Version"),
				color.HiWhiteString(strings.TrimPrefix(getVersion().KubecmVersion, "v")))
			fmt.Printf("%s: %s\n",
				color.BlueString("GoOs"),
				color.HiWhiteString(getVersion().GoOs))
			fmt.Printf("%s: %s\n",
				color.BlueString("GoArch"),
				color.HiWhiteString(getVersion().GoArch))
			newRelease := <-updateMessageChan
			if newRelease != nil {
				fmt.Printf("\n\n%s %s → %s\n",
					color.YellowString("A new release of kubecm is available:"),
					color.CyanString(strings.TrimPrefix(kubecmVersion, "v")),
					color.GreenString(strings.TrimPrefix(newRelease.Version, "v")))
				if isUnderHomebrew() {
					fmt.Printf("To upgrade, run: %s\n", "brew update && brew upgrade kubecm")
				}
				fmt.Printf("%s\n\n",
					termlink.ColorLink("Click into the release page", newRelease.URL, "yellow"))
				//color.YellowString(newRelease.URL))
			}
		},
	}
}

// getVersion returns version.
func getVersion() version {
	return version{
		v.Version,
		v.GoOs,
		v.GoArch,
	}
}

// Check whether the gh binary was found under the Homebrew prefix
func isUnderHomebrew() bool {
	brewExe, err := safeexec.LookPath("brew")
	if err != nil {
		return false
	}

	brewPrefixBytes, err := exec.Command(brewExe, "--prefix").Output()
	if err != nil {
		return false
	}

	path, err := exec.LookPath(os.Args[0])
	if err != nil {
		return false
	}
	kubecmBinary, err := filepath.Abs(path)
	if err != nil {
		return false
	}

	brewBinPrefix := filepath.Join(strings.TrimSpace(string(brewPrefixBytes)), "bin") + string(filepath.Separator)
	return strings.HasPrefix(kubecmBinary, brewBinPrefix)
}
