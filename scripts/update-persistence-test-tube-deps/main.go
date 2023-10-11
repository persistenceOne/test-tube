package main

import (
	"fmt"
	"net/http"

	"io/ioutil"
	"log"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

func main() {
	// Get the absolute path of this binary's directory
	dir, err := filepath.Abs(filepath.Dir(os.Args[0]))

	// store args[1] as persistence rev
	persistenceRev := os.Args[1]

	if err != nil {
		log.Fatal(err)
	}

	libpersistencetesttubeModPath := filepath.Join(dir, "../../packages/persistence-test-tube/libpersistencetesttube/go.mod")
	persistenceModUrl := fmt.Sprintf("https://raw.githubusercontent.com/persistenceOne/persistenceCore/%s/go.mod", persistenceRev)

	libpersistencetesttubeMod := readModFromFile(libpersistencetesttubeModPath)
	persistenceMod := readModFromUrl(persistenceModUrl)

	replaceModFileReplaceDirectives(persistenceMod, libpersistencetesttubeMod)
	writeMod(libpersistencetesttubeMod, libpersistencetesttubeModPath)
}

func readModFromUrl(url string) *modfile.File {
	// Download the go.mod file
	resp, err := http.Get(url)
	if err != nil {
		log.Fatal(err)
	}
	defer resp.Body.Close()

	// turn the response body into a bytes
	bytes, err := ioutil.ReadAll(resp.Body)

	// Parse the go.mod file
	f, err := modfile.Parse("go.mod", bytes, nil)
	if err != nil {
		log.Fatal(err)
	}

	return f
}
func readModFromFile(modPath string) *modfile.File {
	// Read the contents of the go.mod file
	bytes, err := ioutil.ReadFile(modPath)
	if err != nil {
		log.Fatal(err)
	}

	// Parse the go.mod file
	f, err := modfile.Parse(modPath, bytes, nil)
	if err != nil {
		log.Fatal(err)
	}

	return f
}

func replaceModFileReplaceDirectives(from, to *modfile.File) {
	fmt.Printf("Drop replace directives for `%s`:\n", to.Module.Mod.Path)

	// Drop all replace directives from `to` go.mod
	for _, rep := range to.Replace {
		fmt.Printf("  - %s %s => %s %s\n", rep.Old.Path, rep.Old.Version, rep.New.Path, rep.New.Version)
		to.DropReplace(rep.Old.Path, rep.Old.Version)
	}

	// Cleanup the go.mod file
	to.Cleanup()

	fmt.Println("---")

	fmt.Printf("Add replace directives for `%s`:\n", to.Module.Mod.Path)

	// Add all replace directives from `from` go.mod
	for _, rep := range from.Replace {
		fmt.Printf("  - %s %s => %s %s\n", rep.Old.Path, "", rep.New.Path, rep.New.Version)
		to.AddReplace(rep.Old.Path, "", rep.New.Path, rep.New.Version)
	}

	// Sort the blocks
	to.SortBlocks()
}

func writeMod(mod *modfile.File, modPath string) {
	// Write the go.mod file
	content, err := mod.Format()
	if err != nil {
		log.Fatal(err)
	}

	err = ioutil.WriteFile(modPath, content, 0644)
	if err != nil {
		log.Fatal(err)
	}
}
