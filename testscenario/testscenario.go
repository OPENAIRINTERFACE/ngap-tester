package testscenario

import (
	"bufio"
	"bytes"
	"encoding/json"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/factory"
	"github.com/omec-project/gnbsim/gnodeb"
	"github.com/omec-project/gnbsim/gnodeb/context"
	"github.com/omec-project/gnbsim/logger"
	simgnbctx "github.com/openairinterface/ngap-tester/simgnb/context"
	"github.com/openairinterface/ngap-tester/simue"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
	"github.com/sirupsen/logrus"
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
	Id             string
	Description    string
	Status         StatusType
	Action         func(test *TestScenario) error
	UePassedCount  uint
	UeFailedCount  uint
	GnbFailedCount uint
	SimUe          map[string]*simuectx.SimUe
	ErrorList      []error
	SimGnb         []*simgnbctx.SimGnb
	WaitGroups     []sync.WaitGroup

	ReadChan chan common.InterfaceMessage // gnb AMF to Scenario

	/* logger */
	Log *logrus.Entry
}

func ListOfTestFromFile(filename string) []string {
	var testList []string
	if filename == "" {
		return testList
	}
	readFile, err := os.Open(filename)
	if err != nil {
		log.Fatalf("could not open test-file %s", filename)
		return testList
	}
	fileScanner := bufio.NewScanner(readFile)
	fileScanner.Split(bufio.ScanLines)
	for fileScanner.Scan() {
		testList = append(testList, fileScanner.Text())
	}
	err = readFile.Close()
	if err != nil {
		log.Fatalf("could not close test-file %s", filename)
		return testList
	}
	return testList
}

func CheckIfTestIsInList(testName string, testList []string) bool {
	for _, value := range testList {
		if value == testName {
			return true
		}
	}
	return false
}

func CreateTestSuite(c *cli.Context) []TestScenario {
	var testSuite []TestScenario
	run_all := c.Bool("all")
	run_random := c.Bool("random")
	testName := c.String("one-test")
	testFile := c.String("test-file")
	testsList := ListOfTestFromFile(testFile)

	testTestName := "TC1"
	if run_all || run_random || testName == testTestName || CheckIfTestIsInList(testTestName, testsList) {
		scenario := TestScenario{
			Id:          testTestName,
			Description: "UE Initiated Registration Procedures - SUCIas id (UE and AMF Interactions- NAS) - Single gNB",
			Status:      SCENARIO_NOT_RUN,
			Action:      runScenarioTC1,
			ReadChan:    make(chan common.InterfaceMessage, 9),
			SimUe:       make(map[string]*simuectx.SimUe),
			Log:         logger.ScenarioLog.WithField(logger.FieldScenario, testTestName),
		}
		testSuite = append(testSuite, scenario)
	}
	testTestName = "TC1a"
	if run_all || run_random || testName == testTestName || CheckIfTestIsInList(testTestName, testsList) {
		scenario := TestScenario{
			Id:          testTestName,
			Description: "Loop SUCI Registration with Single UE",
			Status:      SCENARIO_NOT_RUN,
			Action:      runScenarioTC1a,
			ReadChan:    make(chan common.InterfaceMessage, 9),
			SimUe:       make(map[string]*simuectx.SimUe),
			Log:         logger.ScenarioLog.WithField(logger.FieldScenario, testTestName),
		}
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
	var wg sync.WaitGroup
	var Mu sync.Mutex

	gnbSims, err := InitializeAllGnbSims()
	if err != nil {
		logger.AppLog.Errorln("Failed to initialize gNodeBs: ", err)
		return err
	}

	var status bool = true

	for i, tst := range ts {
		tst.SimGnb = gnbSims
		err := tst.ProvisionUes()

		if err != nil {
			tst.Log.Errorln("Failed to provision Ues: ", err)
			return err
		}

		wg.Add(1)
		go func(scn *TestScenario) {
			defer wg.Done()
			err := scn.Action(scn)
			ts[i].Status = scn.Status
			// Execution for the UE is complete. Count UE result as success or failure
			if err != nil {
				Mu.Lock()
				status = false
				Mu.Unlock()
			}
		}(&tst)

		if factory.AppConfig.Configuration.ExecScenariosInParallel == false {
			logger.AppLog.Traceln("Waiting for scenario ", tst.Id, " to continue")
			wg.Wait()
		}
	}
	if factory.AppConfig.Configuration.ExecScenariosInParallel == true {
		logger.AppLog.Traceln("Waiting for for all scenarios to finish processing...")
		wg.Wait()
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

func InitializeAllGnbSims() ([]*simgnbctx.SimGnb, error) {
	gnbs := factory.AppConfig.Configuration.Gnbs
	simGnbs := make([]*simgnbctx.SimGnb, len(factory.AppConfig.Configuration.Gnbs))
	i := 0
	for _, gnb := range gnbs {
		simGnb := simgnbctx.NewSimGnb()
		err := gnodeb.Init(gnb, simGnb.ReadChan)
		if err != nil {
			gnb.Log.Errorln("Failed to initialize GNodeB, err:", err)
			return nil, err
		}
		err = simGnb.Init(gnb.GnbName)
		if err != nil {
			gnb.Log.Errorln("Failed to initialize SimGnb, err:", err)
			return nil, err
		}
		simGnbs[i] = simGnb
		i = i + 1
	}
	return simGnbs, nil
}
func (test *TestScenario) AllocateSimUes(gnb *context.GNodeB) error {

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
	return nil
}

func (test *TestScenario) ProvisionUes() error {
	// Allocate objects separatly from launch of scenarios
	// May help reduce delays between start of scenarios in seq or //
	keysUeProf := make([]string, 0, len(factory.AppConfig.Configuration.UeProfiles))
	for k := range factory.AppConfig.Configuration.UeProfiles {
		keysUeProf = append(keysUeProf, k)
	}
	sort.Strings(keysUeProf)

	client := &http.Client{}

	for k := 0; k < len(keysUeProf); k = k + 1 {
		ueProfile := factory.AppConfig.Configuration.UeProfiles[keysUeProf[k]]
		if ueProfile.Provision.CreateSubscriber {
			test.Log.Infoln("Provisioning ", ueProfile.NumUes, " subscribers for ", keysUeProf[k], " UE profile...")
			// Unmarshall JSON like string from config file into GO struct
			jBlob := []byte(ueProfile.Provision.CreateJsonContent)
			var subscriberProvision simuectx.SubscriberProvision
			err := json.Unmarshal(jBlob, &subscriberProvision)
			if err != nil {
				test.Log.Errorln("ProvisionUes failed to handle JSON content:", err)
				return err
			}
			startImsi, err := strconv.Atoi(ueProfile.StartImsi)
			if err != nil {
				err = fmt.Errorf("invalid imsi value: %v", ueProfile.StartImsi)
				test.Log.Errorln(err)
				return err
			}
			for count, imsi := 1, startImsi; count <= ueProfile.NumUes; count, imsi = count+1, imsi+1 {
				imsiStr := "imsi-" + strconv.Itoa(imsi)
				err = test.ProvisionWithJson(imsiStr, ueProfile.Provision.CreateRestUrl, ueProfile.Provision.DeleteRestUrl, &subscriberProvision, client)
				if err != nil {
					test.Log.Errorln("Failed to provision subscriber ", imsiStr, ": ", err)
					return err
				}
				test.Log.Traceln("Provisioned subscriber ", imsiStr)
			}
			test.Log.Infoln("Provisioning ", ueProfile.NumUes, " subscribers for ", keysUeProf[k], " UE profile, done")

		}
	}
	return nil
}

// TODO may be moved at the rigth place when several UE profiles will be implemented
func (test *TestScenario) ProvisionWithJson(imsiStr string, createRestUrl string, deleteRestUrl string, provision *simuectx.SubscriberProvision, httpClient *http.Client) error {

	provision.Supi = imsiStr
	createRestUrl = strings.Replace(createRestUrl, "SUBSCRIBER-SUPI", imsiStr, -1)
	deleteRestUrl = strings.Replace(deleteRestUrl, "SUBSCRIBER-SUPI", imsiStr, -1)
	payload, err := json.Marshal(provision)
	if err != nil {
		return err
	}

	request, err := http.NewRequest(http.MethodDelete, deleteRestUrl, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err := httpClient.Do(request)
	if err != nil {
		return err
	}

	if response.StatusCode != http.StatusNoContent && response.StatusCode != http.StatusAccepted {
		err = fmt.Errorf("Non-Deleted HTTP status: %v", response.StatusCode)
		return err
	}
	response.Body.Close()

	request, err = http.NewRequest(http.MethodPut, createRestUrl, bytes.NewBuffer(payload))
	if err != nil {
		return err
	}
	request.Header.Set("Content-Type", "application/json; charset=utf-8")

	response, err = httpClient.Do(request)
	if err != nil {
		return err
	}

	defer response.Body.Close()

	if response.StatusCode != http.StatusCreated {
		err = fmt.Errorf("Non-Created HTTP status: %v", response.StatusCode)
		return err
	}
	return nil
}

func (scnr *TestScenario) InitImsi(gnb *context.GNodeB, imsiStr string, ueModel string) error {
	simUe := simue.InitUE(imsiStr, ueModel, gnb)
	scnr.SimUe[imsiStr] = simUe
	return nil
}

func (scnr *TestScenario) SendEventToSimUe(imsiStr string, event common.EventType) {
	msg := &common.UeMessage{}
	msg.Event = event
	scnr.SimUe[imsiStr].ReadChan <- msg
}
