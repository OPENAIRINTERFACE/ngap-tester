package simue

import (
	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/realue"
	"github.com/omec-project/nas"
	"github.com/omec-project/ngap/ngapType"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

func PerformRegisterProcedure(simUe *simuectx.SimUe) (common.InterfaceMessage, error) {
	var err error
	simUe.Log.Traceln("PerformRegisterProcedure")
	//-------------------
	// Step send NAS Registration_Request
	//-------------------
	err = SendRegisterRequest(simUe)
	if err != nil {
		return nil, err
	}

	//-------------------
	// Step expect receive NGAP-DownlinkNASTransport/NAS-Authentication_Request
	//-------------------
	n1N2MsgResp, n1DecodedMsg, err := ExpectReceiveN1N2(simUe,
		ngapType.ProcedureCodeDownlinkNASTransport,
		nas.MsgTypeAuthenticationRequest, 3)
	if err != nil {
		return n1N2MsgResp, err
	}

	//-------------------
	// Step answer to NGAP-DownlinkNASTransport/NAS-Authentication_Request
	//-------------------
	for {
		err = PerformAuthenticationProcedure(simUe, n1DecodedMsg)
		if err != nil {
			// LG: TODO or not TODO here
			// Handling sync failures, etc
			return n1N2MsgResp, err
		} else {
			break
		}
	}
	n1N2MsgResp, n1DecodedMsg, err = ExpectReceiveN1N2(simUe,
		ngapType.ProcedureCodeDownlinkNASTransport,
		nas.MsgTypeSecurityModeCommand, 3)
	if err != nil {
		return n1N2MsgResp, err
	}

	//-------------------
	// Step answer to NGAP-DownlinkNASTransport/NAS-Security_Mode_Command
	//-------------------
	err = PerformSecurityProcedure(simUe, n1DecodedMsg)
	if err != nil {
		return n1N2MsgResp, err
	}

	//-------------------
	// Step expect receive  NGAP-InitialContextSetup/NAS-Register_Accept
	// LG: TODO a func dedicated to this
	//-------------------
	n1N2MsgResp, n1DecodedMsg, err = ExpectReceiveN1N2(simUe,
		ngapType.ProcedureCodeInitialContextSetup,
		nas.MsgTypeRegistrationAccept, 3)
	if err != nil {
		return n1N2MsgResp, err
	}

	// Process IEs
	err = realue.HandleRegistrationAccept(simUe.RealUe, n1DecodedMsg.RegistrationAccept)
	if err != nil {
		return nil, err
	}
	//-------------------
	// Step Send back  NGAP-InitialContextSetupResponse/NAS-none
	//-------------------
	err = SendInitialContextSetupResponse(simUe)
	if err != nil {
		return nil, err
	}

	//-------------------
	// Step Send back  NGAP-DownlinkNASTransport/NAS-Registration_Complete
	//-------------------
	err = SendRegistrationComplete(simUe)
	return nil, err
}
