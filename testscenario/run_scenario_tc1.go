package testscenario

import (
	"log"
)

func runScenarioTC1(test *TestScenario) error {
	log.Printf("Running scenario TC1: %s", test.Description)
	test.Status = SCENARIO_FAILED
	return nil
}
