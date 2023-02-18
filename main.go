package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/cloudrecipes/packagejson"
	"github.com/rivo/tview"
)

func main() {
	doPrintHelp := flag.Bool("h", false, "Print help and exit")
	flag.Parse()

	if *doPrintHelp == true {
		printUsageAndExit()
	}

	values := flag.Args()

	if len(values) < 2 {
		printUsageAndExit()
	}

	lastParam := values[len(values)-1]
	workdir, err := filepath.Abs(lastParam)
	if err != nil {
		log.Fatal("Error: cannot resolve ", workdir, ": ", err)
	}
	fmt.Println("Working dir:", workdir)
	file, err := os.Open(workdir)
	if err != nil {
		log.Fatal("Could not open ", workdir, ": ", err)
	}

	fileInfo, err := file.Stat()
	if !fileInfo.IsDir() {
		log.Fatal("Not a directory ", workdir)
	}
	fmt.Println("Searching for scripts in", workdir)
	var scripts []string
	err = filepath.Walk(workdir,
		func(path string, info os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if info.IsDir() && (strings.HasPrefix(info.Name(), ".") || strings.HasPrefix(info.Name(), "node_modules") || strings.HasPrefix(info.Name(), "test")) {
				return filepath.SkipDir
			}
			if info.Name() == "package.json" {
				fmt.Println("Found", path)
				scripts = append(scripts, path)
			}
			return nil
		})
	if err != nil {
		log.Fatal("Could not iterate discover scripts ", ": ", err)
	}
	fmt.Println("Found", len(scripts), "scripts total")
	patterns := values[:len(values)-1]
	fmt.Println("Detecting patterns", patterns)
	packagesAndPatterns := make(map[string][]string)
	for _, path := range scripts {
		content, err := os.ReadFile(path)

		if err != nil {
			log.Fatal("Could not read", path, ": ", err)
		}

		pkg, err := packagejson.Parse(content)
		if err != nil {
			log.Fatal("Could not parse ", path, ": ", err)
		}
		err = pkg.Validate()
		if err != nil {
			log.Fatal("Not a valid package.json ", path, ": ", err)
		}
		var foundScripts []string
		for name := range pkg.Scripts {
			for _, pattern := range patterns {
				if pattern == name {
					fmt.Println("Found matching script", name, "in", path)
					foundScripts = append(foundScripts, name)
				}
			}
		}
		if len(foundScripts) > 0 {
			packagesAndPatterns[path] = foundScripts
		}
	}
	fmt.Println("Found scripts", packagesAndPatterns)
	app := tview.NewApplication()
	flex := tview.NewFlex()
	for file, scripts := range packagesAndPatterns {
		for _, script := range scripts {
			rel, err := filepath.Rel(workdir, file)
			if err != nil {
				log.Fatal("Could not relativize ", file, " to ", workdir, ": ", err)
			}
			text := tview.NewTextView().SetWordWrap(true)
			text.SetBorder(true).SetTitle("┤ " + script + " | " + rel + " ├")
			text.SetChangedFunc(func() {
				app.Draw()
			})

			c := exec.Command("npm", "run", script)
			c.Dir = filepath.Dir(file)
			stdout, err := c.StdoutPipe()
			if err != nil {
				log.Fatal("Could not get stdout for npm run ", script, " for ", file)
			}
			stderr, err := c.StderrPipe()
			if err != nil {
				log.Fatal("Could not get stderr for npm run ", script, " for ", file)
			}
			err = c.Start()
			if err != nil {
				log.Fatal("Could start npm run ", script, " for ", file)
			}
			scanOut := bufio.NewScanner(stdout)
			go updateDisplay(*scanOut, text)
			scanErr := bufio.NewScanner(stderr)
			go updateDisplay(*scanErr, text)
			flex.AddItem(text, 0, 1, false)
		}
	}
	if err := app.SetRoot(flex, true).SetFocus(flex).Run(); err != nil {
		panic(err)
	}
}

func updateDisplay(scanner bufio.Scanner, text *tview.TextView) {
	scanner.Split(bufio.ScanRunes)
	for scanner.Scan() {
		text.Write(scanner.Bytes())
	}
}

func printUsageAndExit() {
	fmt.Println("Usage: aggro <pattern-1> ... <pattern-n> <basedir>")
	fmt.Println("    <pattern> - a name for a scripts block in package.json (eg. 'watch:local')")
	fmt.Println("    <basedir> - directory inside which to look for package.json files")
	os.Exit(0)
}
