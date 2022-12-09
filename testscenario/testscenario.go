package testscenario

import (
	"bufio"
	"fmt"
	"log"
	"math/rand"
	"net"
	"os"
	"sync"
	"time"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/factory"
	"github.com/omec-project/gnbsim/gnodeb"
	"github.com/omec-project/gnbsim/gnodeb/context"
	gnbctx "github.com/omec-project/gnbsim/gnodeb/context"
	"github.com/omec-project/gnbsim/logger"
	"github.com/omec-project/gnbsim/simue"
	simuectx "github.com/omec-project/gnbsim/simue/context"
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

type ScenarioUeContext struct {
	TrigEventsChan chan *common.InterfaceMessage // Receiving Events from the REST interface
	WriteSimChan   chan common.InterfaceMessage  // Sending events to SIMUE -  start proc and proc parameters
	ReadChan       chan *common.InterfaceMessage // simUe to profile ?
	WriteGnbUeChan chan common.InterfaceMessage  // Sending events to gnb

	SimUe      *simuectx.SimUe
	CurrentGnb *gnbctx.GNodeB

	/* logger */
	Log *logrus.Entry
}

type TestScenario struct {
	Id            string
	Description   string
	Status        StatusType
	Action        func(test *TestScenario) error
	UePassedCount uint
	UeFailedCount uint
	SimUe         map[string]*ScenarioUeContext
	ErrorList     []error

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
			SimUe:       make(map[string]*ScenarioUeContext),
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
			SimUe:       make(map[string]*ScenarioUeContext),
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

	err := gnodeb.InitializeAllGnbs()
	if err != nil {
		logger.AppLog.Errorln("Failed to initialize gNodeBs: ", err)
		return err
	}

	var status bool = true

	for i, tst := range ts {
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

func PerformNgapSetupProcedure(test *TestScenario, gnbName string, amfName string) error {
	var err error

	gnbCtx, err := factory.AppConfig.Configuration.GetGNodeB(gnbName)
	if err != nil {
		test.Log.Errorln("GetGNodeB returned:", err)
		return err
	}

	if gnbCtx.Amf == nil {
		amf, err := factory.AppConfig.Configuration.GetAmf(amfName)
		if err != nil {
			test.Log.Errorln("GetAmf returned:", err)
			return err
		}
		if amf.AmfIp == "" {
			// It is important to do this lookup just in time, not at simulation startup
			addrs, err := net.LookupHost(amf.AmfHostName)
			if err != nil {
				return fmt.Errorf("failed to resolve amf host name: %v, err: %s", amf.AmfHostName, err)
			}
			gnbCtx.Amf = gnbctx.NewGnbAmf(addrs[0], gnbctx.NGAP_SCTP_PORT)
		}
	}

	err = gnbCtx.CpTransport.ConnectToPeer(gnbCtx.Amf)
	if err != nil {
		test.Log.Errorln("ConnectToAmf returned:", err)
		return err
	}

	successFulOutcome, err := gnodeb.PerformNgSetup(gnbCtx, gnbCtx.Amf)
	if err != nil {
		test.Log.Errorln("PerformNGSetup returned:", err)
	} else if !successFulOutcome {
		err = fmt.Errorf("Result: FAIL, Expected SuccessfulOutcome, received UnsuccessfulOutcome")
	}
	return err
}

func (scnr *TestScenario) InitImsi(gnb *context.GNodeB, imsiStr string) {
	readChan := make(chan *common.InterfaceMessage)
	simUe := simue.InitUE(imsiStr, "default", gnb, readChan)
	scenarioUeContext := ScenarioUeContext{WriteSimChan: simUe.ReadChan}
	scenarioUeContext.ReadChan = readChan
	trigChan := make(chan *common.InterfaceMessage)
	scenarioUeContext.TrigEventsChan = trigChan
	scenarioUeContext.Log = logger.ScnrUeCtxLog.WithField(logger.FieldSupi, imsiStr)
	scenarioUeContext.SimUe = simUe
	scenarioUeContext.WriteGnbUeChan = simUe.WriteGnbUeChan
	scnr.SimUe[imsiStr] = &scenarioUeContext
}

func (scnr *TestScenario) SendEventToSimUe(imsiStr string, event common.EventType) {
	msg := &common.UeMessage{}
	msg.Event = event
	scnr.SimUe[imsiStr].WriteSimChan <- msg
}

func (scnr *TestScenario) SendUserDataPacket(imsiStr string) {
	scnr.Log.Infoln("Initiating User Data Packet Generation Procedure")
	msg := &common.UeMessage{}
	// TODO
	msg.UserDataPktCount = 10
	// TODO msg.DefaultAs = ue.ProfileCtx.DefaultAs
	msg.Event = common.DATA_PKT_GEN_REQUEST_EVENT

	/* TODO: Solve timing issue. Currently UE may start sending user data
	 * before gnb has successfuly sent PDU Session Resource Setup Response
	 * or before 5g core has processed it
	 */
	//ue.Log.Infoln("Please wait, initiating uplink user data in 3 seconds ...")
	//time.Sleep(3 * time.Second)

	scnr.SimUe[imsiStr].WriteSimChan <- msg
}

func (scnr_ue *ScenarioUeContext) SendToGnbUe(msg common.InterfaceMessage) {
	scnr_ue.Log.Traceln("Sending", msg.GetEventType(), "to GnbUe")
	scnr_ue.WriteGnbUeChan <- msg
}

func (scnr_ue *ScenarioUeContext) HandleEvents() {
	for msg := range scnr_ue.ReadChan {
		event := (*msg).GetEventType()
		scnr_ue.Log.Infoln("Handling event:", event)

	}
	return
}
