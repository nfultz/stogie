package main

import (
	"flag"
    "log"
	"github.com/nfultz/stogie/version"
    "container/list"
)



func debug(level int, format string, args ...interface{}) {
    if flagvar.verbose >= level {
        log.Printf(format, args)
    }


    return
}

type stogieFlags struct {

    version bool
    help bool

    simulate bool
    nofolding bool

    verbose int

    dir string
    target string

    adopt bool

    ignoreRegex string
    deferRegex string
    overrideRegex string

    stowpkgs []string

}

var flagvar stogieFlags


var adds, dels = list.New(), list.New()

func main() {


	flag.BoolVar(&flagvar.version, "version", false, "Show version number")
    flag.BoolVar(&flagvar.help, "help",false, "Show this help")

	flag.BoolVar(&flagvar.simulate, "simulate", false, "Do not actually make any file system changes")
	flag.BoolVar(&flagvar.nofolding, "no-folding", false,
      "Disable folding of newly stowed directories when stowing, and refolding of newly foldable directories when unstowing.")

    flag.IntVar(&flagvar.verbose, "verbose", 0, "Set verbosity level: 0, 1, 2, 3, and 4; 0 is the default.")

	flag.StringVar(&flagvar.dir, "d",".", "Set stow dir (default is current dir)")
	flag.StringVar(&flagvar.target, "t","", "Set target dir (default is parent of stow dir)")

	flag.BoolVar(&flagvar.adopt, "adopt", false, "(Use with care) Import existing files into stow package from target")

    // Conflict resolution
	flag.StringVar(&flagvar.ignoreRegex, "ignore","", "Ignore files matching this regex.")
	flag.StringVar(&flagvar.deferRegex, "defer","", "Don't stow files matching this regex if the file is already stowed to another package.")
	flag.StringVar(&flagvar.overrideRegex, "override","", "Force stowing files matching this regex if the file is already stowed to another package.")

	flag.Parse()
    flagvar.stowpkgs = flag.Args()

	if flagvar.version {
        version.PrintVersion()
		return
	}

    if flagvar.help {
        flag.PrintDefaults()
        return
    }

    toadd, todel := true, false
    for _, s := range flagvar.stowpkgs {
        switch s {
        case "-S":
            toadd, todel = true, false
        case "-D":
            toadd, todel = false, true
        case "-R":
            toadd, todel = true, true
        default:
            if todel {
                dels.PushBack(s)
            }
            if toadd {
                adds.PushBack(s)
            }

        }
    }

    for e := dels.Front(); e != nil; e = e.Next() {
        // do something with e.Value
        log.Printf("Del \t%s\n", e.Value)
    }
    for e := adds.Front(); e != nil; e = e.Next() {
        // do something with e.Value
        log.Printf("ADD \t%s\n", e.Value)
    }
}
