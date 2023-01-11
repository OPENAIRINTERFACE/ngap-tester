package simue

import (
	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/nas"
	"github.com/omec-project/ngap/ngapType"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

func PerformDeregisterProcedureUEOriginatingDeregistration(
	simUe *simuectx.SimUe,
) (common.InterfaceMessage, error) {
	var err error
	simUe.Log.Traceln("PerformDeregisterProcedureUEOriginatingDeregistration")
	//-------------------
	// Step send NAS Deregistration_Request
	//-------------------
	err = SendDeregisterRequest(simUe)
	if err != nil {
		return nil, err
	}

	//-------------------
	// Step expect receive NGAP-DownlinkNASTransport/NAS-Deregistration-Accept
	//-------------------
	n1N2MsgResp, _, err := ExpectReceiveN1N2(simUe,
		ngapType.ProcedureCodeDownlinkNASTransport,
		nas.MsgTypeDeregistrationAcceptUEOriginatingDeregistration, 3)

	//-------------------
	// Step expect receive NGAP-UEContextReleaseRequest
	//-------------------
	n1N2MsgResp, _, err = ExpectReceiveN1N2(simUe,
		ngapType.ProcedureCodeUEContextRelease,
		0, 3)

	//-------------------
	// Step send NGAP UEContextReleaseComplete
	//-------------------
	err = PerformUEContextReleaseProcedureNwInitiated(
		simUe,
		n1N2MsgResp.NgapPdu,
		0, 0, false,
	)
	if err != nil {
		return nil, err
	}

	return n1N2MsgResp, err
}
