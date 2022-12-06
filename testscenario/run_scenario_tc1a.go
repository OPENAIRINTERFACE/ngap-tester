package testscenario

func runScenarioTC1a(test *TestScenario) error {
	test.Log.Infoln("Running scenario ", test.Id, " : ", test.Description)
	test.Status = SCENARIO_PASSED
	return nil
}
