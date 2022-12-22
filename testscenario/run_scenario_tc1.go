package testscenario

import (
	"fmt"
	"sync"

	"github.com/omec-project/gnbsim/factory"
	"github.com/openairinterface/ngap-tester/simgnb"
	simgnbctx "github.com/openairinterface/ngap-tester/simgnb/context"
	"github.com/openairinterface/ngap-tester/simue"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

const UE_PROFILE string = "default" // Should be listed in yaml config file
const FIRST_AMF string = "amf1"     // Should be listed in yaml config file
const FIRST_GNB_POS int = 0         // value is index in sorted list of gnb names

const (
	TC1_WG_NG_SETUP = iota
	TC1_WG_SIM_GNB_END
	TC1_WG_SIM_UE_END
	TC1_WG_LAST
)

func runScenarioTC1(test *TestScenario) error {
	var Mu sync.Mutex
	test.Log.Infoln("Running scenario ", test.Id, " : ", test.Description)
	test.Status = SCENARIO_UNDEFINED

	// Array of waitgroup
	test.WaitGroups = make([]sync.WaitGroup, TC1_WG_LAST)

	// ===================================================================
	// Internals: Allocate objects separatly from launch of scenarios
	// May help reduce delays between start of scenarios in seq or //
	// ===================================================================
	// Actually start attaching UE on first gNB, TODO configure this later for HO for example
	err := test.AllocateSimUes(test.SimGnb[0].GnB)
	if err != nil {
		err = fmt.Errorf("failed to allocate SimUe(s): %v", err)
		test.Log.Errorln(err)
		return err
	}

	for i := 0; i < len(test.SimGnb); i = i + 1 {
		test.WaitGroups[TC1_WG_NG_SETUP].Add(1)
		test.WaitGroups[TC1_WG_SIM_GNB_END].Add(1)
		go func(simGnb *simgnbctx.SimGnb) {
			defer test.WaitGroups[TC1_WG_SIM_GNB_END].Done()
			err := test.runScenarioTC1Gnb(simGnb)
			// Execution for the gNB is complete. Count gNB result as success or failure
			Mu.Lock()
			if err != nil {
				test.GnbFailedCount++
				test.ErrorList = append(test.ErrorList, err)
				test.Status = SCENARIO_FAILED
			}
			Mu.Unlock()
		}(test.SimGnb[i])
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
		test.WaitGroups[TC1_WG_SIM_UE_END].Add(1)
		scnUeCtx := test.SimUe[imsiStr]
		go func(simUe *simuectx.SimUe) {
			defer test.WaitGroups[TC1_WG_SIM_UE_END].Done()
			err := test.runScenarioTC1Ue(simUe, imsiStr)
			// Execution for the UE is complete. Count UE result as success or failure
			Mu.Lock()
			if err != nil {
				test.UeFailedCount++
				test.ErrorList = append(test.ErrorList, err)
				test.Status = SCENARIO_FAILED
			} else {
				test.UePassedCount++
			}
			Mu.Unlock()
		}(scnUeCtx)

		if !factory.AppConfig.Configuration.ExecUesInParallel {
			test.Log.Traceln("Waiting for UE ", imsiStr, " to continue...")
			test.WaitGroups[TC1_WG_SIM_UE_END].Wait()
		}
	}
	if factory.AppConfig.Configuration.ExecUesInParallel {
		test.Log.Infoln("Waiting for all UEs to finish processing...")
		test.WaitGroups[TC1_WG_SIM_UE_END].Wait()
	}
	test.Log.Infoln("Waiting for all gNB to finish processing...")
	test.WaitGroups[TC1_WG_SIM_GNB_END].Wait()
	if test.Status == SCENARIO_FAILED {
		test.Log.Infoln("Scenario ended with errors")
	} else {
		test.Status = SCENARIO_PASSED
		test.Log.Infoln("Scenario ended with success")
	}

	return nil
}

func (test *TestScenario) runScenarioTC1Ue(simUe *simuectx.SimUe, imsiStr string) error {
	// ===================================================================
	// SCENARIO concerning UEs starts here
	// ===================================================================
	test.Log.Traceln("runScenarioTC1Ue started ")
	test.WaitGroups[TC1_WG_NG_SETUP].Wait()

	_, err := simue.PerformRegisterProcedure(simUe)
	if err != nil {
		test.Log.Errorln("runScenarioTC1Ue ended with error: ", err)
		return err
	}
	_, err = simue.PerformDeregisterProcedureUEOriginatingDeregistration(simUe)

	if err != nil {
		test.Log.Errorln("runScenarioTC1Ue ended with error: ", err)
	} else {
		test.Log.Traceln("runScenarioTC1Ue ended")
	}
	return err
}

func (test *TestScenario) runScenarioTC1Gnb(simGnb *simgnbctx.SimGnb) error {
	// ===================================================================
	// SCENARIO concerning gNB starts here
	// ===================================================================
	test.Log.Traceln("runScenarioTC1Gnb started ")

	_, err := simgnb.PerformNgapSetupProcedure(simGnb)
	if err != nil {
		err = fmt.Errorf("Failed to Perform NGAP Setup Procedure: %v", err)
		test.Log.Errorln(err)
	}
	test.WaitGroups[TC1_WG_NG_SETUP].Done()
	test.WaitGroups[TC1_WG_SIM_UE_END].Wait()
	test.Log.Traceln("runScenarioTC1Gnb ended")

	return err
}
