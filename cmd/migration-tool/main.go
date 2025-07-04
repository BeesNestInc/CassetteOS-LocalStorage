package main

import (
	_ "embed"
	"flag"
	"fmt"
	"os"

	interfaces "github.com/BeesNestInc/CassetteOS-Common"
	"github.com/BeesNestInc/CassetteOS-Common/utils/systemctl"
	"github.com/BeesNestInc/CassetteOS-LocalStorage/common"
)

const (
	localStorageConfigDirPath  = "/etc/cassetteos"
	localStorageConfigFilePath = "/etc/cassetteos/local-storage.conf"
	localStorageName           = "cassetteos-local-storage.service"
	localStorageNameShort      = "local-storage"
)

//go:embedded ../../build/sysroot/etc/casstteos/local-storage.conf.sample
// var _localStorageConfigFileSample string

var (
	commit = "private build"
	date   = "private build"

	_logger *Logger
)

// var _status *version.GlobalMigrationStatus

func main() {
	versionFlag := flag.Bool("v", false, "version")
	debugFlag := flag.Bool("d", true, "debug")
	forceFlag := flag.Bool("f", false, "force")
	flag.Parse()

	if *versionFlag {
		fmt.Printf("v%s\n", common.Version)
		os.Exit(0)
	}

	println("git commit:", commit)
	println("build date:", date)

	_logger = NewLogger()

	if os.Getuid() != 0 {
		_logger.Info("Root privileges are required to run this program.")
		os.Exit(1)
	}

	if *debugFlag {
		_logger.DebugMode = true
	}

	if !*forceFlag {
		isRunning, err := systemctl.IsServiceRunning(localStorageName)
		if err != nil {
			_logger.Error("Failed to check if %s is running", localStorageName)
			panic(err)
		}

		if isRunning {
			_logger.Info("%s is running. If migration is still needed, try with -f.", localStorageName)
			os.Exit(1)
		}
	}

	migrationTools := []interfaces.MigrationTool{
		NewMigrationDummy(),
	}

	var selectedMigrationTool interfaces.MigrationTool

	// look for the right migration tool matching current version
	for _, tool := range migrationTools {
		migrationNeeded, err := tool.IsMigrationNeeded()
		if err != nil {
			panic(err)
		}

		if migrationNeeded {
			selectedMigrationTool = tool
			break
		}
	}

	if selectedMigrationTool == nil {
		_logger.Info("No migration to proceed.")
		return
	}

	if err := selectedMigrationTool.PreMigrate(); err != nil {
		panic(err)
	}

	if err := selectedMigrationTool.Migrate(); err != nil {
		panic(err)
	}

	if err := selectedMigrationTool.PostMigrate(); err != nil {
		_logger.Error("Migration succeeded, but post-migration failed: %s", err)
	}
}
