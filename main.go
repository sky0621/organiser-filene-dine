package main

import (
	"path/filepath"
)

const (
	operationAll          = 9
	operationPrepare      = 1
	operationCreateOutDir = 2
	operationCopy         = 3
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
}
