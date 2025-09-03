// Copyright 2025 Sonic Labs
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

package config

import (
	"flag"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/urfave/cli/v2"
)

func TestGetFlagValue(t *testing.T) {
	// app for testing
	app := cli.NewApp()
	app.Commands = []*cli.Command{
		{
			Name: "testcmd",
			Flags: []cli.Flag{
				&cli.IntFlag{
					Name: "intflag",
				},
				&cli.Uint64Flag{
					Name: "uint64flag",
				},
				&cli.Int64Flag{
					Name: "int64flag",
				},
				&cli.StringFlag{
					Name: "stringflag",
				},
				&cli.PathFlag{
					Name: "pathflag",
				},
				&cli.BoolFlag{
					Name: "boolflag",
				},
				&cli.StringSliceFlag{
					Name: "stringsliceflag",
				},
			},
		},
	}

	// Setup test cases
	testCases := []struct {
		name          string
		setupFlags    func() (*cli.Context, error)
		flagToTest    interface{}
		expectedValue interface{}
	}{
		{
			name: "IntFlag value",
			setupFlags: func() (*cli.Context, error) {
				set := flag.NewFlagSet("test", 0)
				set.Int("intflag", 42, "")
				ctx := cli.NewContext(app, set, nil)
				ctx.Command = app.Commands[0]
				return ctx, nil
			},
			flagToTest:    cli.IntFlag{Name: "intflag"},
			expectedValue: 42,
		},
		{
			name: "Uint64Flag value",
			setupFlags: func() (*cli.Context, error) {
				set := flag.NewFlagSet("test", 0)
				set.Uint64("uint64flag", 100, "")
				ctx := cli.NewContext(app, set, nil)
				ctx.Command = app.Commands[0]
				return ctx, nil
			},
			flagToTest:    cli.Uint64Flag{Name: "uint64flag"},
			expectedValue: uint64(100),
		},
		{
			name: "Int64Flag value",
			setupFlags: func() (*cli.Context, error) {
				set := flag.NewFlagSet("test", 0)
				set.Int64("int64flag", 200, "")
				ctx := cli.NewContext(app, set, nil)
				ctx.Command = app.Commands[0]
				return ctx, nil
			},
			flagToTest:    cli.Int64Flag{Name: "int64flag"},
			expectedValue: int64(200),
		},
		{
			name: "StringFlag value",
			setupFlags: func() (*cli.Context, error) {
				set := flag.NewFlagSet("test", 0)
				set.String("stringflag", "test-string", "")
				ctx := cli.NewContext(app, set, nil)
				ctx.Command = app.Commands[0]
				return ctx, nil
			},
			flagToTest:    cli.StringFlag{Name: "stringflag"},
			expectedValue: "test-string",
		},
		{
			name: "PathFlag value",
			setupFlags: func() (*cli.Context, error) {
				set := flag.NewFlagSet("test", 0)
				set.String("pathflag", "/test/path", "")
				ctx := cli.NewContext(app, set, nil)
				ctx.Command = app.Commands[0]
				return ctx, nil
			},
			flagToTest:    cli.PathFlag{Name: "pathflag"},
			expectedValue: "/test/path",
		},
		{
			name: "BoolFlag value",
			setupFlags: func() (*cli.Context, error) {
				set := flag.NewFlagSet("test", 0)
				set.Bool("boolflag", true, "")
				ctx := cli.NewContext(app, set, nil)
				ctx.Command = app.Commands[0]
				return ctx, nil
			},
			flagToTest:    cli.BoolFlag{Name: "boolflag"},
			expectedValue: true,
		},
		{
			name: "StringSliceFlag value",
			setupFlags: func() (*cli.Context, error) {
				set := flag.NewFlagSet("test", 0)
				set.Var(cli.NewStringSlice("value1", "value2"), "stringsliceflag", "")
				ctx := cli.NewContext(app, set, nil)
				ctx.Command = app.Commands[0]
				return ctx, nil
			},
			flagToTest:    cli.StringSliceFlag{Name: "stringsliceflag"},
			expectedValue: []string{"value1", "value2"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			ctx, err := tc.setupFlags()
			assert.NoError(t, err)

			value := getFlagValue(ctx, tc.flagToTest)
			assert.Equal(t, tc.expectedValue, value)
		})
	}
}
