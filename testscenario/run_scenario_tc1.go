package testscenario

import (
	"fmt"
	"sort"
	"strconv"
	"sync"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/factory"
	"github.com/omec-project/gnbsim/realue"
)

const UE_PROFILE string = "default" // Should be listed in yaml config file
const FIRST_AMF string = "amf1"     // Should be listed in yaml config file
const FIRST_GNB_POS int = 0         // value is index in sorted list of gnb names

func runScenarioTC1(test *TestScenario) error {
	var wg sync.WaitGroup
	var Mu sync.Mutex
	test.Log.Infoln("Running scenario ", test.Id, " : ", test.Description)
	test.Status = SCENARIO_FAILED

	gnb, err := factory.AppConfig.Configuration.GetGNodeBAt(FIRST_GNB_POS)
	if err != nil {
		err = fmt.Errorf("Failed to fetch gNB context: %v", err)
		test.Log.Errorln(err)
		return err
	}

	// Test scenario
	// Actually this procedure is synchronous
	err = PerformNgapSetupProcedure(test, gnb.GnbName, FIRST_AMF)
	if err != nil {
		err = fmt.Errorf("Failed to Perform NGAP Setup Procedure: %v", err)
		test.Log.Errorln(err)
		return err
	}

	// Allocate objects separatly from launch of scenarios
	// May help reduce delays between start of scenarios in seq or //
	keysUeProf := make([]string, 0, len(factory.AppConfig.Configuration.UeProfiles))
	for k := range factory.AppConfig.Configuration.UeProfiles {
		test.Log.Traceln("key UE profile ", k)
		keysUeProf = append(keysUeProf, k)
	}
	sort.Strings(keysUeProf)
	for k := 0; k < len(keysUeProf); k = k + 1 {
		ueProfile := factory.AppConfig.Configuration.UeProfiles[keysUeProf[k]]
		startImsi, err := strconv.Atoi(ueProfile.StartImsi)
		if err != nil {
			err = fmt.Errorf("invalid imsi value: %v", ueProfile.StartImsi)
			test.Log.Errorln(err)
			return err
		}
		for count, imsi := 1, startImsi; count <= ueProfile.NumUes; count, imsi = count+1, imsi+1 {
			imsiStr := "imsi-" + strconv.Itoa(imsi)
			test.InitImsi(gnb, imsiStr, keysUeProf[k])
			test.Log.Traceln("provision UE ", imsiStr)
		}
	}

	keysImsi := make([]string, 0, len(test.SimUe))
	for k := range test.SimUe {
		keysImsi = append(keysImsi, k)
	}
	for i := 0; i < len(keysImsi); i = i + 1 {
		imsiStr := keysImsi[i]
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

func (scn *TestScenario) runScenarioTC1Ue(scnrUeCtx *ScenarioUeContext, imsiStr string) error {

	scn.Log.Traceln("runScenarioTC1Ue started ")
	nasPdu, err := realue.HandleRegRequestEvent(scnrUeCtx.SimUe.RealUe, nil)
	if err != nil {
		return err
	}
	msg := realue.FormUuMessage(common.N1_SEND_SDU_EVENT+common.NAS_5GMM_REGISTRATION_REQUEST, nasPdu)
	msg.Tac = scnrUeCtx.SimUe.GnB.SupportedTaList[0].Tac
	msg.NrCgi = scnrUeCtx.SimUe.GnB.NrCgiCellList[0]
	scnrUeCtx.SendToGnbUe(msg)

	scnrUeCtx.HandleEvents()
	scn.Log.Traceln("runScenarioTC1Ue ended")
	return nil
}
