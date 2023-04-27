package flags

import "github.com/urfave/cli/v2"

var (
	// APIRecordingSrcFileFlag defines path to data recorded on API
	APIRecordingSrcFileFlag = cli.PathFlag{
		Name:    "api-recording",
		Usage:   "Path to source file with recorded API data",
		Aliases: []string{"r"},
	}

	// WorkersFlag defines number of threads for execution
	WorkersFlag = cli.IntFlag{
		Name: "workers",
		Usage: "defines the thread number for api-replay. " +
			"The exact value is used for number of Executor threads, " +
			"number of Comparator threads is the number divided by 2 since the Execution is much slower;" +
			"default: 4",
		Value:   4,
		Aliases: []string{"w"},
	}

	// ContinueOnFailure does not stop the program when unmatched results are found
	ContinueOnFailure = cli.BoolFlag{
		Name:    "continue-on-failure",
		Usage:   "if used, any mismatched data will not exit the program",
		Aliases: []string{"cof"},
	}
)
