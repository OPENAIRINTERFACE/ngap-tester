package testscenario

import (
	"fmt"
	"sync"

	"github.com/omec-project/gnbsim/factory"
	"github.com/openairinterface/ngap-tester/simue"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

const UE_PROFILE string = "default" // Should be listed in yaml config file
const FIRST_AMF string = "amf1"     // Should be listed in yaml config file
const FIRST_GNB_POS int = 0         // value is index in sorted list of gnb names

func runScenarioTC1(test *TestScenario) error {
	var wg sync.WaitGroup
	var Mu sync.Mutex
	test.Log.Infoln("Running scenario ", test.Id, " : ", test.Description)
	test.Status = SCENARIO_FAILED

	// ===================================================================
	// SCENARIO Item here concerning gNB interface management
	// ===================================================================
	gnb, err := factory.AppConfig.Configuration.GetGNodeBAt(FIRST_GNB_POS)
	if err != nil {
		err = fmt.Errorf("Failed to fetch gNB context: %v", err)
		test.Log.Errorln(err)
		return err
	}
	_, err = PerformNgapSetupProcedure(test, gnb.GnbName, FIRST_AMF)
	if err != nil {
		err = fmt.Errorf("Failed to Perform NGAP Setup Procedure: %v", err)
		test.Log.Errorln(err)
		return err
	}
	// ===================================================================
	// Internals: Allocate objects separatly from launch of scenarios
	// May help reduce delays between start of scenarios in seq or //
	// ===================================================================
	err = test.AllocateSimUes(gnb)
	if err != nil {
		err = fmt.Errorf("Failed to allocate SimUe(s): %v", err)
		test.Log.Errorln(err)
		return err
	}

	// ===================================================================
	// Internals: go routines logic concerning UE scenario
	// ===================================================================
	keysImsi := make([]string, 0, len(test.SimUe))
	for k := range test.SimUe {
		keysImsi = append(keysImsi, k)
	}
	for i := 0; i < len(keysImsi); i = i + 1 {
		imsiStr := keysImsi[i]
		wg.Add(1)
		scnUeCtx := test.SimUe[imsiStr]
		go func(simUe *simuectx.SimUe) {
			defer wg.Done()
			err := test.runScenarioTC1Ue(simUe, imsiStr)
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
			test.Log.Traceln("Waiting for UE ", imsiStr, " to continue...")
			wg.Wait()
		}
	}
	if factory.AppConfig.Configuration.ExecUesInParallel == true {
		test.Log.Infoln("Waiting for all UEs to finish processing...")
		wg.Wait()
	}
	test.Log.Infoln("Scenario ended")
	test.Status = SCENARIO_PASSED
	return nil
}

func (scn *TestScenario) runScenarioTC1Ue(simUe *simuectx.SimUe, imsiStr string) error {
	// ===================================================================
	// SCENARIO concerning UEs starts here
	// ===================================================================
	scn.Log.Traceln("runScenarioTC1Ue started ")
	_, err := simue.PerformRegisterProcedure(simUe)
	scn.Log.Traceln("runScenarioTC1Ue ended")
	return err
}
