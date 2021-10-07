package version

import (
	"fmt"
	"os"
)

func ShowVersion(version string, gitCommit string, gitBranch string, buildTime string) {
	fmt.Println("version:", version)
	fmt.Println("git commit:", gitCommit)
	fmt.Println("git branch:", gitBranch)
	fmt.Println("built time:", buildTime)
	os.Exit(0)
}
