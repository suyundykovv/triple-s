package flags

import (
	"flag"
	"fmt"
	"log"
	"os"
	"strings"
)

var (
	Port           string
	StorageDir     string
	restrictedDirs = []string{"flags", "handlers", "models", "servers", "storage", "utils", "../", "./"}
)

func isRestrictedDir(dir string) bool {
	for _, restricted := range restrictedDirs {
		if strings.EqualFold(dir, restricted) {
			return true
		}
	}
	return false
}

func Setup() {
	defaultPort := "6000"
	defaultStorageDir := "./data"
	help := flag.Bool("help", false, "Display help information")
	flag.StringVar(&Port, "port", defaultPort, "Port to run the server on")
	flag.StringVar(&StorageDir, "directory", defaultStorageDir, "Directory for file storage")
	flag.Parse()

	if isRestrictedDir(StorageDir) {
		log.Fatalf("The specified directory '%s' is restricted. Please choose a different name.", StorageDir)
	}
	if *help {
		printUsage()
		os.Exit(0)
	}
	if err := os.MkdirAll(StorageDir, os.ModePerm); err != nil {
		log.Fatalf("Failed to create storage directory %s: %v", StorageDir, err)
	}
	log.Printf("Using storage directory: %s", StorageDir)
}

func printUsage() {
	fmt.Println("Simple Storage Service.")
	fmt.Println()
	fmt.Println("Usage:")
	fmt.Println("    triple-s [-port <N>] [-directory <S>]")
	fmt.Println("    triple-s --help")
	fmt.Println()
	fmt.Println("Options:")
	fmt.Println("  --help     Show this screen.")
	fmt.Println("  --port N   Port number")
	fmt.Println("  --dir S    Path to the directory")
}
