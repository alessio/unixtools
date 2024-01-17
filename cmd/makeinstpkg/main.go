package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"

	"al.essio.dev/pkg/tools/internal/instpkg"
	"al.essio.dev/pkg/tools/internal/version"
)

var (
	helpMode      bool
	versionMode   bool
	verboseMode   bool
	configFile    string
	defaultConfig bool

	info *log.Logger
)

func init() {
	log.SetFlags(0)
	log.SetPrefix("makeinstpkg: ")
	log.SetOutput(os.Stderr)

	info = log.New(io.Discard, "INFO: makeinstpkg: ", 0)

	flag.BoolVar(&helpMode, "help", false, "display this help and exit.")
	flag.BoolVar(&versionMode, "version", false, "output version information and exit.")
	flag.BoolVar(&verboseMode, "V", false, "print verbose output.")
	flag.BoolVar(&defaultConfig, "C", false, "print the default configuration and exit.")
	flag.StringVar(&configFile, "c", ".makeinstpkg", "path to the configuration file.")

	flag.Usage = usage
	flag.CommandLine.SetOutput(os.Stderr)
}

func main() {
	var config instpkg.Configuration

	flag.Parse()
	handleHelpAndVersionModes()

	if verboseMode {
		info.SetOutput(os.Stderr)
	}

	if defaultConfig {
		bs, err := json.MarshalIndent(instpkg.DefaultConfiguration(), "", "\t")
		if err != nil {
			log.Fatal(err)
		}

		fmt.Println(string(bs))
		os.Exit(0)
	}

	wd, err := os.Getwd()
	if err != nil {
		log.Fatalf("getcwd: %v", err)
	}

	info.Println("reading config file", configFile)
	cfgBs, err := os.ReadFile(configFile)
	if err != nil {
		log.Fatalf("couldn't open %s: %v", configFile, err)
	}

	if err := json.Unmarshal(cfgBs, &config); err != nil {
		log.Fatalf("couldn't read configuration: %v", err)
	}

	if err := config.Validate(); err != nil {
		log.Fatalf("invalid configuratoon: %v", err)
	}

	// Create the actual root directory.
	rootDir, err := os.MkdirTemp(wd, ".mkinstpkg-")
	if err != nil {
		log.Fatalf("mkdirtemp: %v", err)
	}

	info.Println("created temporary directory", rootDir)

	appTopDir := fmt.Sprintf("%s/%s", rootDir, config.Package.Name)
	appDir := fmt.Sprintf("%s/%s", appTopDir, config.Package.Version)
	binDir := fmt.Sprintf("%s/bin", appDir)
	if err := os.MkdirAll(binDir, 0755); err != nil {
		log.Fatalf("mkdirall: %v", err)
	}
	info.Println("created directory", binDir)

	// Fix directories permissions.
	dirs := []string{rootDir, appTopDir, appDir, binDir}
	for _, d := range dirs {
		info.Println(" fixing permissions of the directory", d)
		if err := os.Chmod(d, 0755); err != nil {
			log.Fatalf("couldn't chmod the directory %s: %v", d, err)
		}
	}

	files, err := os.ReadDir(config.SourceDir)
	if err != nil {
		log.Fatalf("couldn't read directory %s: %v", config.SourceDir, err)
	}

	for _, f := range files {
		dstFile := filepath.Join(binDir, f.Name())

		info.Println("installing file", dstFile)
		if err := installFile(filepath.Join(config.SourceDir, f.Name()), dstFile); err != nil {
			log.Fatal(err)
		}

		if config.Signing.Identity != "" && !config.Signing.SkipCode {
			info.Println("signing file", dstFile)
			if err := runCommand("codesign", "-s", config.Signing.Identity,
				"--options=runtime", dstFile); err != nil {
				log.Fatal(err)
			}
		}
	}

	// Handle the uninstaller script
	uninstallerFilename := filepath.Join(appDir, "uninstall.sh")

	if uninstallerFile, err := os.Create(uninstallerFilename); err != nil {
		log.Fatalf("couldn't create the uninstaller file %s: %v", uninstallerFilename, err)
	} else {
		uninstallSh := template.Must(template.New("uninstall").Parse(instpkg.FlatBinDirUninstall))
		if err := uninstallSh.Execute(uninstallerFile, config.Package); err != nil {
			log.Fatal(err)
		}

		if err := uninstallerFile.Chmod(0755); err != nil {
			log.Fatalf("couldn't chmod %s: %v", uninstallerFile, err)
		}

		uninstallerFile.Close()
	}

	// Build the .pkg file
	args := []string{"--root", rootDir,
		"--install-location", config.InstallLocation,
		"--identifier", config.Package.Identifier,
		"--version", config.Package.Version,
	}

	if config.ScriptsDir != "" {
		args = append(args, "--scripts", config.ScriptsDir)
	}

	args = append(args, config.Package.Name+".pkg")

	if err := runCommand("pkgbuild", args...); err != nil {
		log.Fatal(err)
	}

	// Sign the .pkg file
	if !config.Signing.SkipInstaller {

	}

	info.Println("removing the temporary directory", rootDir)
	if err := os.RemoveAll(rootDir); err != nil {
		log.Printf("couldn't remove the diredtory %s: %v", rootDir, err)
	}
}

func handleHelpAndVersionModes() {
	if !helpMode && !versionMode {
		return
	}

	switch {
	case helpMode:
		usage()
	case versionMode:
		version.PrintWithCopyright()
	}

	os.Exit(0)
}

func usage() {
	s := `Usage: makeinstpkg [-c FILE]
Build a MacOS pkg file from a flat directory
structure containing only command-line utilities.

It expects a configuration file in the current directory
named '.makeinstpkg'. Use the '-c' flag to set a different
path for the configuration file.

Options:`
	_, _ = fmt.Fprintln(os.Stderr, s)
	flag.PrintDefaults()
}

const bufSize = 8192

func installFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}

	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	_, err = os.Stat(dst)
	if err == nil {
		return fmt.Errorf("file %s already exists", dst)
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}

	buf := make([]byte, bufSize)
	for {
		n, err := source.Read(buf)
		if err != nil && err != io.EOF {
			return err
		}

		if n == 0 {
			break
		}

		if _, err := destination.Write(buf[:n]); err != nil {
			return err
		}
	}

	if err := destination.Close(); err != nil {
		return fmt.Errorf("couldn't close the file %s: %v", dst, err)
	}

	if err := os.Chmod(dst, 0755); err != nil {
		return fmt.Errorf("couldn't chmod the file %s: %v", dst, err)
	}

	return nil
}

func runCommand(name string, args ...string) error {
	info.Println("running", name, args)
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}
