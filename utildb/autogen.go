// Copyright 2024 Fantom Foundation
// This file is part of Aida Testing Infrastructure for Sonic
//
// Aida is free software: you can redistribute it and/or modify
// it under the terms of the GNU Lesser General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Aida is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU Lesser General Public License for more details.
//
// You should have received a copy of the GNU Lesser General Public License
// along with Aida. If not, see <http://www.gnu.org/licenses/>.

package utildb

import (
	"errors"
	"fmt"
	"github.com/0xsoniclabs/aida/config"
	"io/fs"
	"os"
	"time"

	"github.com/0xsoniclabs/substate/db"
	"github.com/urfave/cli/v2"
)

// SetLock creates lockfile in case of error while generating
func SetLock(cfg *config.Config, message string) error {
	lockFile := cfg.AidaDb + ".autogen.lock"

	// Write the string to the file
	err := os.WriteFile(lockFile, []byte(message), 0655)
	if err != nil {
		return fmt.Errorf("error writing to lock file %v; %v", lockFile, err)
	} else {
		return nil
	}
}

// GetLock checks existence and contents of lockfile
func GetLock(cfg *config.Config) (string, error) {
	lockFile := cfg.AidaDb + ".autogen.lock"

	// Read lockfile contents
	content, err := os.ReadFile(lockFile)
	if errors.Is(err, fs.ErrNotExist) {
		return "", nil
	} else if err != nil {
		return "", fmt.Errorf("error reading from file; %v", err)
	}

	return string(content), nil
}

// AutogenRun is used to record/update aida-db
func AutogenRun(cfg *config.Config, g *Generator) error {
	g.Log.Noticef("Starting substate generation %d - %d", g.Opera.FirstEpoch, g.TargetEpoch)

	start := time.Now()
	// stop opera to be able to export events
	errCh := startOperaRecording(g.Cfg, g.TargetEpoch)

	// wait for opera recording response
	err, ok := <-errCh
	if ok && err != nil {
		return err
	}
	g.Log.Noticef("Recording (%v) for epoch range %d - %d finished. It took: %v", g.Cfg.ClientDb, g.Opera.FirstEpoch, g.TargetEpoch, time.Since(start).Round(1*time.Second))
	g.Log.Noticef("Total elapsed time: %v", time.Since(g.start).Round(1*time.Second))

	// reopen aida-db
	g.AidaDb, err = db.NewDefaultBaseDB(cfg.AidaDb)
	if err != nil {
		return fmt.Errorf("cannot create new db; %v", err)
	}

	err = g.Opera.getOperaBlockAndEpoch(false)
	if err != nil {
		return err
	}

	return g.Generate()
}
