package main

import (
	"flag"
    "log"
	"github.com/nfultz/stogie/version"
    "os"
    "path"
    "path/filepath"
	"strings"
)



func debug(level int, format string, args ...interface{}) {
    if flagvar.verbose >= level {
        log.Printf(format, args...)
    }


    return
}

func die(format string, args ...interface{}) {
	log.Printf(format, args...)

    os.Exit(1)
}

func fpabs(path string) string {
	ret, _ := filepath.Abs(path)
	return ret
}

func addDots(file string, link string) string {
	debug(9, "adddots %s %s", file, link)
	d,f := path.Split(file)
	if d == "" {
		return link
	}
    return addDots(f, filepath.Join("..", link) )
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

type Task interface {
    Print()
	Run()
}

type LinkTask struct {
    link string
    target string
}

type UnlinkTask struct {
    file string
}


type MkdirTask struct {
    dir string
}

type AdoptTask struct {
	pkg string
    file string
}

func (t LinkTask) Print() {
	debug(4, "ln -s %s %s", t.link, t.target)
}

func (t UnlinkTask) Print() {
	debug(4, "rm %s", t.file)
}

func (t MkdirTask) Print() {
	debug(4, "mkdir -p %s", t.dir)
}

func (t LinkTask) Run() {
	os.Symlink(t.link, t.target)
}

func (t UnlinkTask) Run() {
	os.Remove(t.file)
}

func (t MkdirTask) Run() {
	os.Mkdir(t.dir, os.ModePerm)
}

func main() {

    // Part 1 deal with flags

	flag.BoolVar(&flagvar.version, "version", false, "Show version number")
    flag.BoolVar(&flagvar.help, "help",false, "Show this help")

	flag.BoolVar(&flagvar.simulate, "simulate", false, "Do not actually make any file system changes")
	flag.BoolVar(&flagvar.nofolding, "no-folding", true,
      "Disable folding of newly stowed directories when stowing, and refolding of newly foldable directories when unstowing. (Default true, false not implemented)")

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

	if flagvar.target == "" {
		flagvar.target = filepath.Join(flagvar.dir, "..")

	}

    // part 1b - list of package verbs to two lists;
    adds, dels := make([]string, 0, 100), make([]string, 0, 100)
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

			if !strings.HasSuffix(s, string(os.PathSeparator)) {
               s = s + string(os.PathSeparator)
			}

            if todel {
                dels = append(dels, s)
            }
            if toadd {
                adds = append(adds, s)
            }

        }
    }





    for _, e := range dels{
        // do something with e.Value
        debug(3, "DEL \t%s\n", e)
    }
    for _, e := range adds{
        // do something with e.Value
        debug(3, "ADD \t%s\n", e)
    }

	stowDirRel := strings.TrimPrefix(fpabs(flagvar.dir), fpabs(flagvar.target) + string(os.PathSeparator))
	debug(3, "Stowdir relative to target is (%s)", stowDirRel)




    tasks := make([]Task, 0)

	// Plan unstow of each pkg
    for _, pkg := range dels{
		pkgdir := filepath.Join(flagvar.dir, pkg)
		fi, _ := os.Lstat(pkgdir)
		if fi == nil || ! fi.IsDir() {
			die("The stow directory '%s' does not contain pkg '%s'.", flagvar.dir, pkg)
		}

		debug(2, "Planning unstow of %s", pkg)
		filepath.Walk(pkgdir, func(path string, info os.FileInfo, err error) error {
			if info.IsDir() {
				return nil
			}

			file := strings.TrimPrefix(path, pkg)
			target := filepath.Join(flagvar.target, file)
			link := filepath.Join(stowDirRel, pkg, file )
			link = addDots(file, link)

			rl, _ := os.Readlink(target)
			debug(3, "**ReadLink:(%s) -> (%s)", target,  rl)


			if rl == link {
				debug(3, "**Queuing unlink task for %s", target)
				tasks = append(tasks, UnlinkTask{file:target})
			}

			return nil
		})
		debug(2, "Planning unstow of %s done", pkg)
    }

    for _, pkg := range adds {
		pkgdir := filepath.Join(flagvar.dir, pkg)
		fi, _ := os.Lstat(pkgdir)
		if fi == nil || ! fi.IsDir() {
			die("The stow directory '%s' does not contain pkg '%s'.", flagvar.dir, pkg)
		}

		debug(2, "Planning stow of %s", pkg)
		filepath.Walk(pkgdir, func(path string, info os.FileInfo, err error) error {
			if path == pkgdir {
            	return nil
			}
			file := strings.TrimPrefix(path, pkg)
			target := filepath.Join(flagvar.target, file)

			if info.IsDir() {
				debug(3, "**Mkdir:(%s)", target)
				tasks = append(tasks, MkdirTask{dir: target})
				return nil
			}

			link := filepath.Join(stowDirRel, pkg, file )
			link = addDots(file, link)

			// Prepend with .. for subfolders

			debug(3, "**CreateLink:(%s) -> (%s)", target,  link)


			tasks = append(tasks, LinkTask{target:target, link:link})

			return nil
		})
		debug(2, "Planning stow of %s done", pkg)
    }

	for _, e := range tasks {
		e.Print()
	}


	for _, e := range tasks {
		e.Run()
	}


}
