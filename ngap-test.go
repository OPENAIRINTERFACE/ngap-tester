package main

import (
	"errors"
	"log"
	"os"

	"github.com/openairinterface/ngap-tester/testscenario"
	"github.com/urfave/cli"
)

func main() {
	cli.VersionFlag = cli.BoolFlag{
		Name:  "print-version, V",
		Usage: "print only the version",
	}
	cli.AppHelpTemplate = `NAME:
   {{.Name}}
USAGE:
   {{.HelpName}} {{if .Commands}}command{{end}} {{if .VisibleFlags}}[command options]{{end}}
{{if .Description}}
DESCRIPTION:
   {{.Description}}{{end}}{{if len .Authors}}
AUTHOR:
   {{range .Authors}}{{ . }}{{end}}
   {{end}}
{{if .VisibleCommands}}
COMMANDS:{{range .VisibleCategories}}{{if .Name}}
   {{.Name}}:{{range .VisibleCommands}}
     {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{else}}{{range .VisibleCommands}}
   {{join .Names ", "}}{{"\t"}}{{.Usage}}{{end}}{{end}}{{end}}{{end}}
{{if .VisibleFlags}}
GLOBAL OPTIONS:
   {{range .VisibleFlags}}{{.}}
   {{end}}{{end}}{{if .Copyright }}
COPYRIGHT:
   {{.Copyright}}
   {{end}}{{if .Version}}
VERSION:
   {{.Version}}
   {{end}}
`
	app := cli.NewApp()
	app.UseShortOptionHandling = true
	app.Name = "NGAP and 5G-NAS Tester"
	app.Description = "An RAN emulator with a test-suite"
	app.Version = "0.0.1"
	app.Commands = []cli.Command{
		{
			Name:   "run",
			Usage:  "Runs the specified test-suite",
			Flags:  getCliFlags(),
			Action: run_tests,
		},
		{
			Name:   "list",
			Usage:  "Lists the specified test-suite",
			Flags:  getCliFlags(),
			Action: list_tests,
		},
	}

	err := app.Run(os.Args)
	if err != nil {
		log.Fatal(err)
	}
}

func run_tests(c *cli.Context) error {
	err := check_flags(c)
	if err != nil {
		return err
	}
	testArray := testscenario.CreateTestSuite(c)
	err = testscenario.RunTestsuite(testArray)
	if err != nil {
		return err
	}

	return nil
}

func list_tests(c *cli.Context) error {
	err := check_flags(c)
	if err != nil {
		return err
	}
	testArray := testscenario.CreateTestSuite(c)
	testscenario.DisplayTestsuite(testArray)
	return nil
}

func check_flags(c *cli.Context) error {
	run_all := c.Bool("all")
	run_random := c.Bool("random")
	testFile := c.String("test-file")
	testName := c.String("one-test")

	// Global Options are exclusive.
	var nbOptions uint8 = 0
	if run_all {
		nbOptions++
	}
	if run_random {
		nbOptions++
	}
	if testFile != "" {
		nbOptions++
		if _, err := os.Stat(testFile); errors.Is(err, os.ErrNotExist) {
			return cli.NewExitError("The test file does not exist", 5)
		}
	}
	if testName != "" {
		nbOptions++
	}

	if nbOptions > 1 {
		return cli.NewExitError("Global Options are exclusive", 2)
	} else if nbOptions == 0 {
		return cli.NewExitError("Please specify one (1) global option", 3)
	}
	return nil
}

func getCliFlags() []cli.Flag {
	return []cli.Flag{
		cli.BoolFlag{
			Name:  "all, a",
			Usage: "will run all the tests in the testsuite",
		},
		cli.BoolFlag{
			Name:  "random, r",
			Usage: "will run all the tests in the testsuite in random order",
		},
		cli.StringFlag{
			Name:  "test-file, tf",
			Usage: "`TEST-FILE` that contains list of test(s) to run",
		},
		cli.StringFlag{
			Name:  "one-test, o",
			Usage: "will run the specified `TEST-NAME` only",
		},
	}
}
