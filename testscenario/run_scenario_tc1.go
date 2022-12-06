package testscenario

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/omec-project/gnbsim/factory"
)

const UE_PROFILE string = "default" // Should be listed in yaml config file
const NUM_UES int = 2
const FIRST_GNB string = "gnb1" // Should be listed in yaml config file
const FIRST_AMF string = "amf1" // Should be listed in yaml config file

func runScenarioTC1(test *TestScenario) error {
	var wg sync.WaitGroup
	var Mu sync.Mutex
	test.Log.Infoln("Running scenario %s: %s", test.Id, test.Description)
	test.Status = SCENARIO_FAILED

	ueProfile, err := factory.AppConfig.Configuration.GetUeProfile(UE_PROFILE)
	if err != nil {
		test.Log.Errorln("%s", err)
		return err
	}
	startImsi, err := strconv.Atoi(ueProfile.StartImsi)
	if err != nil {
		err = fmt.Errorf("invalid imsi value:%v", ueProfile.StartImsi)
		test.Log.Errorln("%s", err)
		return err
	}

	gnb, err := factory.AppConfig.Configuration.GetGNodeB(FIRST_GNB)
	if err != nil {
		err = fmt.Errorf("Failed to fetch gNB context: %v", err)
		test.Log.Errorln("%s", err)
		return err
	}

	// Test scenario
	err = PerformNgapSetupProcedure(test, FIRST_GNB, FIRST_AMF)
	if err != nil {
		err = fmt.Errorf("Failed to Perform NGAP Setup Procedure: %v", err)
		test.Log.Errorln("%s", err)
		return err
	}

	// Allocate objects separatly from launch of scenarios
	// May help reduce delays between start of scenarios in seq or //
	for count, imsi := 1, startImsi; count <= NUM_UES; count, imsi = count+1, imsi+1 {
		imsiStr := "imsi-" + strconv.Itoa(imsi)
		test.InitImsi(gnb, imsiStr)
	}

	for count, imsi := 1, startImsi; count <= NUM_UES; count, imsi = count+1, imsi+1 {
		imsiStr := "imsi-" + strconv.Itoa(imsi)
		wg.Add(1)
		scnUeCtx := test.SimUe[imsiStr]
		go func(scnrUeCtx *ScenarioUeContext) {
			defer wg.Done()
			err := test.runScenarioTC1Ue(scnrUeCtx, imsiStr)
			// Execution for the UE is complete. Count UE result as success or failure
			Mu.Lock()
			if err != nil {
				test.UeFailedCount++
				test.ErrorList = append(test.ErrorList, err)
			} else {
				test.UePassedCount++
			}
			Mu.Unlock()
		}(scnUeCtx)

		if factory.AppConfig.Configuration.ExecUesInParallel == false {
			test.Log.Traceln("Waiting for UE %s to continue...", imsiStr)
			wg.Wait()
		}
	}
	if factory.AppConfig.Configuration.ExecUesInParallel == true {
		test.Log.Infoln("Waiting for for all UEs to finish processing...")
		wg.Wait()
	}
	test.Log.Infoln("Scenario ended")
	test.Status = SCENARIO_PASSED
	return nil
}

func (scn *TestScenario) runScenarioTC1Ue(scnrUeCtx *ScenarioUeContext, imsiStr string) error {

	scn.Log.Traceln("runScenarioTC1Ue started ")

	scn.Log.Traceln("runScenarioTC1Ue ended")
	return nil
}
