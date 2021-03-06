package version

import (
        "runtime"
        "fmt"
)

// The git commit that was compiled. This will be filled in by the compiler.
var GitCommit string

// The main version number that is being run at the moment.
const Version = "0.1.0"

var BuildDate = ""

var GoVersion = runtime.Version()

var OsArch = fmt.Sprintf("%s %s", runtime.GOOS, runtime.GOARCH)


func PrintVersion() {
    fmt.Println("Build Date:", BuildDate)
    fmt.Println("Git Commit:", GitCommit)
    fmt.Println("Version:",    Version)
    fmt.Println("Go Version:", GoVersion)
    fmt.Println("OS / Arch:",  OsArch)
    return
}
