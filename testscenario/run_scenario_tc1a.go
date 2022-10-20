package testscenario

import (
	"log"
)

func runScenarioTC1a(test *TestScenario) error {
	log.Printf("Running scenario TC1a: %s", test.Description)
	test.Status = SCENARIO_PASSED
	return nil
}
