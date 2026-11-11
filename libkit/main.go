package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	cmd := os.Args[1]

	switch cmd {
	case "list", "ls":
		cmdList()
	case "search", "find":
		cmdSearch()
	case "info", "show":
		cmdInfo()
	case "test":
		cmdTest()
	case "add":
		cmdAdd()
	case "new", "init":
		cmdNew()
	case "tags", "tag":
		cmdTags()
	case "verify":
		cmdVerify()
	case "version", "-v", "--version":
		fmt.Println("libkit v1.0.0 (Antigravity 2026)")
	case "help", "-h", "--help":
		printUsage()
	default:
		printError(fmt.Sprintf("Unknown command '%s'", cmd))
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	printBanner()
	fmt.Print(`USAGE:
  libkit <command> [arguments...]

COMMANDS:
  list                 List all libraries in the catalog
  search <query>       Search libraries by name, tag, or description
  info <name|id>       Display library SPEC, exported APIs, and README quickstart
  test [name|--all]    Run tests for a library or all discovered Go modules
  add <name> [--to .]  Inject a library into a target Go project's go.mod
  new <project-name>   Scaffold a new Clean Architecture project using this workspace
  tag                  Create v1.1.0 tags for consolidated module roots
  verify               Deep scan all directories and verify 100% test pass rate
  version              Show libkit CLI version

EXAMPLES:
  libkit list
  libkit search jwt
  libkit info go-authkit
  libkit test go-saga
  libkit test --all
  libkit tag
  libkit new my-microservice
`)
}

func cmdList() {
	items, err := LoadCatalog()
	if err != nil {
		printError(fmt.Sprintf("Failed to load catalog: %v", err))
		return
	}

	fmt.Printf("\n📦 %s: %d Production Libraries\n\n", colorize(bold+brightCyan, "C:\\Library Catalog"), len(items))

	headers := []string{"ID", "LIBRARY NAME", "DESCRIPTION", "STDLIB", "STATUS"}
	var rows [][]string

	for _, item := range items {
		rows = append(rows, []string{
			fmt.Sprintf("%03d", item.ID),
			item.Name,
			truncate(item.Description, 50),
			item.StdLib,
			item.Status,
		})
	}

	printTable(headers, rows)
	fmt.Println()
	printInfo(fmt.Sprintf("Run '%s' to view detailed architecture and code snippets.", colorize(cyan, "libkit info <name>")))
}

func cmdSearch() {
	if len(os.Args) < 3 {
		printWarn("Usage: libkit search <query>")
		return
	}

	query := os.Args[2]
	results, err := SearchCatalog(query)
	if err != nil {
		printError(fmt.Sprintf("Search failed: %v", err))
		return
	}

	fmt.Printf("\n🔍 Search results for '%s' (%d found):\n\n", colorize(yellow, query), len(results))
	if len(results) == 0 {
		printWarn("No matching libraries found.")
		return
	}

	headers := []string{"ID", "LIBRARY NAME", "DESCRIPTION", "STATUS"}
	var rows [][]string
	for _, item := range results {
		rows = append(rows, []string{
			fmt.Sprintf("%03d", item.ID),
			item.Name,
			item.Description,
			item.Status,
		})
	}
	printTable(headers, rows)
}

func cmdInfo() {
	if len(os.Args) < 3 {
		printWarn("Usage: libkit info <name|id>")
		return
	}

	target := os.Args[2]
	item, err := GetLibrary(target)
	if err != nil {
		printError(err.Error())
		return
	}

	fmt.Println()
	fmt.Printf("📦 %s (ID: #%03d)\n", colorize(bold+brightCyan, item.Name), item.ID)
	fullModule := fmt.Sprintf("github.com/amarnathdbg101-coder/go/%s", item.Module)
	fmt.Printf("🌐 Remote Module: %s\n", colorize(bold+green, fullModule))
	fmt.Printf("📦 Package Path:  %s\n", colorize(dim, item.ImportPath))
	fmt.Printf("📥 Remote Fetch:  %s\n", colorize(cyan, "go get "+fullModule+"@latest"))
	fmt.Printf("📁 Location:      %s\n", colorize(dim, item.Path))
	fmt.Printf("📝 Description:   %s\n", item.Description)
	fmt.Printf("⚡ Dependencies:  %s\n\n", colorize(green, item.StdLib))

	// Print README
	readme, err := item.ReadDoc("README.md")
	if err == nil {
		fmt.Println(colorize(bold+yellow, "--- QUICKSTART (README.md) ---"))
		fmt.Println(readme)
	}

	// Print SPEC summary
	spec, err := item.ReadDoc("SPEC.md")
	if err == nil {
		fmt.Println(colorize(bold+cyan, "--- ARCHITECTURE SPEC (SPEC.md) ---"))
		fmt.Println(spec)
	}
}

func cmdTest() {
	target := "--all"
	if len(os.Args) >= 3 {
		target = os.Args[2]
	}

	rootDir := FindLibraryRoot()
	if rootDir == "" {
		printError("Library workspace not found; run from the workspace or set LIBKIT_ROOT")
		return
	}

	if target == "--all" || target == "all" {
		cmdVerify()
		return
	}

	item, err := GetLibrary(target)
	if err != nil {
		printError(err.Error())
		return
	}

	printInfo(fmt.Sprintf("Testing %s in %s...", colorize(cyan, item.Name), item.Path))
	runGoTest(item.Path, true)
}

func cmdVerify() {
	rootDir := FindLibraryRoot()
	if rootDir == "" {
		printError("Library workspace not found; run from the workspace or set LIBKIT_ROOT")
		return
	}

	dirs, err := os.ReadDir(rootDir)
	if err != nil {
		printError(fmt.Sprintf("Failed to read root directory: %v", err))
		return
	}

	var libDirs []string
	for _, d := range dirs {
		if d.IsDir() {
			if _, err := os.Stat(filepath.Join(rootDir, d.Name(), "go.mod")); err == nil {
				libDirs = append(libDirs, d.Name())
			}
		}
	}
	sort.Strings(libDirs)

	fmt.Printf("\n🧪 %s across %d libraries in %s...\n\n",
		colorize(bold+brightCyan, "Executing Batch Test Suite"),
		len(libDirs), rootDir)

	passed := 0
	var failed []string
	start := time.Now()

	for i, dname := range libDirs {
		libPath := filepath.Join(rootDir, dname)
		fmt.Printf("[%3d/%3d] Testing %-25s ", i+1, len(libDirs), colorize(cyan, dname))

		ok, out := runGoTest(libPath, false)
		if ok {
			fmt.Printf("%s\n", colorize(green, "PASS"))
			passed++
		} else {
			fmt.Printf("%s\n", colorize(red, "FAIL"))
			failed = append(failed, fmt.Sprintf("%s:\n%s", dname, out))
		}
	}

	elapsed := time.Since(start)
	fmt.Println()
	if len(failed) == 0 {
		printSuccess(fmt.Sprintf("ALL %d LIBRARIES PASSED! (100%% GREEN) in %v", passed, elapsed.Round(time.Millisecond)))
	} else {
		printError(fmt.Sprintf("%d PASSED, %d FAILED in %v", passed, len(failed), elapsed.Round(time.Millisecond)))
		for _, f := range failed {
			fmt.Println(f)
		}
	}
}

func runGoTest(dir string, verbose bool) (bool, string) {
	args := []string{"test", "-count=1"}
	if verbose {
		args = append(args, "-v")
	}
	args = append(args, "./...")

	cmd := exec.Command("go", args...)
	cmd.Dir = dir

	output, err := cmd.CombinedOutput()
	if verbose {
		fmt.Println(string(output))
	}
	return err == nil, string(output)
}

func cmdAdd() {
	if len(os.Args) < 3 {
		printWarn("Usage: libkit add <library-name> [--to projectDir]")
		return
	}

	libName := os.Args[2]
	item, err := GetLibrary(libName)
	if err != nil {
		printError(err.Error())
		return
	}

	targetDir := "."
	for i := 3; i < len(os.Args); i++ {
		if os.Args[i] == "--to" && i+1 < len(os.Args) {
			targetDir = os.Args[i+1]
			break
		}
	}

	goModPath := filepath.Join(targetDir, "go.mod")
	if _, err := os.Stat(goModPath); err != nil {
		printError(fmt.Sprintf("No go.mod found in target directory: %s", targetDir))
		return
	}

	fullModule := fmt.Sprintf("github.com/amarnathdbg101-coder/go/%s", item.Module)
	fullImportPath := item.ImportPath

	// Add replace directive via `go mod edit`
	modEditCmd := exec.Command("go", "mod", "edit",
		"-require", fmt.Sprintf("%s@v0.0.0", fullModule),
		"-replace", fmt.Sprintf("%s=%s", fullModule, filepath.ToSlash(item.ModuleDir)),
	)
	modEditCmd.Dir = targetDir
	out, err := modEditCmd.CombinedOutput()
	if err != nil {
		printError(fmt.Sprintf("Failed to update go.mod: %v\n%s", err, string(out)))
		return
	}

	printSuccess(fmt.Sprintf("Successfully injected %s into %s/go.mod", colorize(bold+green, fullModule), targetDir))
	printInfo(fmt.Sprintf("Import with: import \"%s\"", fullImportPath))
	printInfo(fmt.Sprintf("Or remote install: go get %s@latest", fullModule))
}

func cmdTags() {
	rootDir := FindLibraryRoot()
	if rootDir == "" {
		printError("Library workspace not found; run from the workspace or set LIBKIT_ROOT")
		return
	}

	items, err := LoadCatalog()
	if err != nil {
		printError(fmt.Sprintf("Failed to load catalog: %v", err))
		return
	}

	modules := make(map[string]struct{})
	for _, item := range items {
		if item.Name != item.Module {
			modules[item.Module] = struct{}{}
		}
	}
	printInfo(fmt.Sprintf("Generating consolidated module Git Tags (v1.1.0) across %d modules...", len(modules)))

	created := 0
	moduleNames := make([]string, 0, len(modules))
	for module := range modules {
		moduleNames = append(moduleNames, module)
	}
	sort.Strings(moduleNames)
	for _, module := range moduleNames {
		tagName := fmt.Sprintf("%s/v1.1.0", module)
		// Check if tag already exists
		checkCmd := exec.Command("git", "tag", "-l", tagName)
		checkCmd.Dir = rootDir
		out, _ := checkCmd.Output()
		if strings.TrimSpace(string(out)) == tagName {
			continue
		}

		tagCmd := exec.Command("git", "tag", tagName)
		tagCmd.Dir = rootDir
		if err := tagCmd.Run(); err == nil {
			created++
		}
	}

	printSuccess(fmt.Sprintf("Generated %d new Git tags!", created))
	fmt.Println()
	printInfo("To publish all tags to GitHub, run:")
	fmt.Println(colorize(bold+cyan, "  git push origin --tags"))
	fmt.Println()
	printInfo("After pushing, anyone can install via:")
	fmt.Println(colorize(bold+green, "  go get github.com/amarnathdbg101-coder/go/<module-name>@v1.1.0"))
}

func cmdNew() {
	if len(os.Args) < 3 {
		printWarn("Usage: libkit new <project-name>")
		return
	}

	projectName := os.Args[2]
	targetDir := "."
	if len(os.Args) >= 4 && !strings.HasPrefix(os.Args[3], "-") {
		targetDir = os.Args[3]
	}

	printInfo(fmt.Sprintf("Scaffolding Clean Architecture project '%s' in %s...", colorize(bold+cyan, projectName), targetDir))

	err := ScaffoldProject(targetDir, projectName)
	if err != nil {
		printError(fmt.Sprintf("Scaffold failed: %v", err))
		return
	}

	printSuccess(fmt.Sprintf("Project '%s' created successfully!", projectName))
	fmt.Printf("\nNext steps:\n")
	fmt.Printf("  1. cd %s\n", projectName)
	fmt.Printf("  2. go run cmd/server/main.go\n")
	fmt.Printf("  3. curl http://localhost:8080/health\n\n")
}

func truncate(s string, max int) string {
	if len(s) <= max {
		return s
	}
	return s[:max-3] + "..."
}
