package version

import (
	"fmt"
	"runtime"

	"github.com/spf13/cobra"
)

var (
	version   = "unknown"
	goVersion = fmt.Sprintf("%s %s/%s", runtime.Version(), runtime.GOOS, runtime.GOARCH)
)

var VersionCmd = &cobra.Command{
	Use:   "version",
	Short: "Scargo CLI version",
	Run: func(cmd *cobra.Command, args []string) {
		fmt.Printf("Scargo CLI %s (go: %s)\n", version, goVersion)
	},
}
