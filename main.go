package main

import (
	"path/filepath"
)

const (
	operationAll              = 9
	operationPrepare          = 1
	operationCreateOutDir     = 2
	operationCopy             = 3
	operationCheckDuplication = 4
	operationDeDuplication    = 5
	operationRenameDir        = 6
)

func main() {
	cfg := getConfig()

	createDirectory(filepath.Join(cfg.ToDir, metaDir))

	/****************************************************************
	 * create copy list
	 */
	if cfg.Operation == operationPrepare || cfg.Operation == operationAll {
		listUp(cfg)
	}

	/****************************************************************
	 * create output directory
	 */
	if cfg.Operation == operationCreateOutDir || cfg.Operation == operationAll {
		createOutputDir(cfg)
	}

	/****************************************************************
	 * copy
	 */
	if cfg.Operation == operationCopy || cfg.Operation == operationAll {
		execCopy(cfg.ToDir)
	}

	/****************************************************************
	 * check-duplication
	 */
	if cfg.Operation == operationCheckDuplication {
		checkDuplication(cfg.ToDir)
	}

	/****************************************************************
	 * de-duplication
	 */
	if cfg.Operation == operationDeDuplication {
		deDuplication(cfg.ToDir)
	}

	/****************************************************************
	 * rename-dir
	 */
	if cfg.Operation == operationRenameDir {
		renameDir(cfg.ToDir)
	}
}
