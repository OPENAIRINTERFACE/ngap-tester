package testscenario

import (
	"log"
	"math/rand"
	"time"

	"github.com/urfave/cli"
)

type StatusType uint8

const (
	SCENARIO_NOT_RUN   StatusType = 0
	SCENARIO_PASSED    StatusType = 1
	SCENARIO_FAILED    StatusType = 2
	SCENARIO_UNDEFINED StatusType = 3
)

type TestScenario struct {
	Id          string
	Description string
	Status      StatusType
	Action      func(test *TestScenario) error
}

func CreateTestSuite(c *cli.Context) []TestScenario {
	var testSuite []TestScenario
	run_all := c.Bool("all")
	run_random := c.Bool("random")
	testName := c.String("one-test")
	// TODO : handle the case when a list of scenarios is provided in a file

	if run_all || run_random || testName == "TC1" {
		scenario := TestScenario{
			Id:          "TC1",
			Description: "UE Initiated Registration Procedures - SUCIas id (UE and AMF Interactions- NAS) - Single gNB",
			Status:      SCENARIO_NOT_RUN,
			Action:      runScenarioTC1}
		testSuite = append(testSuite, scenario)
	}
	if run_all || run_random || testName == "TC1a" {
		scenario := TestScenario{
			Id:          "TC1a",
			Description: "Loop SUCI Registration with Single UE",
			Status:      SCENARIO_NOT_RUN,
			Action:      runScenarioTC1a}
		testSuite = append(testSuite, scenario)
	}
	// Shuffle randomly the test-suite
	if run_random {
		rand.Seed(time.Now().UnixNano())
		rand.Shuffle(len(testSuite), func(i, j int) { testSuite[i], testSuite[j] = testSuite[j], testSuite[i] })
	}
	return testSuite
}

func DisplayTestsuite(ts []TestScenario) {
	log.Print("Scenario\t: Description")
	log.Print("------------:-------------------------------------------------------")
	for _, tst := range ts {
		log.Printf("* %s\t: %s", tst.Id, tst.Description)
	}
}

func RunTestsuite(ts []TestScenario) error {
	var status bool = true
	for i, tst := range ts {
		err := tst.Action(&tst)
		ts[i].Status = tst.Status
		if err != nil {
			status = false
		}
	}
	DisplayTestsuiteResults(ts)
	if status {
		return nil
	} else {
		return cli.NewExitError("At least one test case failed", 4)
	}
}

func DisplayTestsuiteResults(ts []TestScenario) {
	log.Print("Scenario\t: Status\t: Description")
	log.Print("------------:-------------------------------------------------------")
	for _, tst := range ts {
		switch tst.Status {
		case SCENARIO_PASSED:
			log.Printf("* %s\t: PASSED\t: %s", tst.Id, tst.Description)
		case SCENARIO_FAILED:
			log.Printf("* %s\t: FAILED\t: %s", tst.Id, tst.Description)
		case SCENARIO_NOT_RUN:
			log.Printf("* %s\t: NOT RUN\t: %s", tst.Id, tst.Description)
		case SCENARIO_UNDEFINED:
			log.Printf("* %s\t: NOT DONE\t: %s", tst.Id, tst.Description)
		default:
			log.Fatalf("* %s\t: UNKNOWN\t: %s", tst.Id, tst.Description)
		}
	}
}
