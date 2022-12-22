package simue

import (
	"errors"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/gnodeb"
	"github.com/omec-project/gnbsim/realue"
	realue_nas "github.com/omec-project/gnbsim/realue/nas"
	"github.com/omec-project/nas"
	"github.com/omec-project/ngap/ngapType"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

func PerformDeregisterProcedureUEOriginatingDeregistration(simUe *simuectx.SimUe) (common.InterfaceMessage, error) {
	var err error
	simUe.Log.Traceln("PerformDeregisterProcedureUEOriginatingDeregistration")
	//-------------------
	// Step send NAS Deregistration_Request
	//-------------------
	nasPdu, err := realue_nas.GetDeregisterRequest(simUe.RealUe)
	if err != nil {
		return nil, err
	}

	sendMsg, err := gnodeb.GetUplinkNASTransport(simUe.GnB, simUe.GnbCpUe, nasPdu)
	if err != nil {
		return nil, err
	}
	msg := FormN2Message(common.N2_SEND_SDU_EVENT, sendMsg)
	SendToGnbUe(simUe, msg)

	//-------------------
	// Step expect receive NGAP-DownlinkNASTransport/NAS-Deregistration-Accept
	//-------------------
	msgResp, ok := simUe.RcvTimedSecondEvent(3)
	if !ok {
		err = errors.New("expected NGAP-DownlinkNASTransport/NAS-Deregistration-Accept coming from AMF timed-out")
		return nil, err
	}
	if msgResp.GetEventType() != common.N1_N2_RECV_SDU_EVENT {
		err = errors.New("unexpected response to N1 Deregistration-Request")
		return msgResp, err
	}
	n1N2MsgResp := msgResp.(*common.N1N2Message)
	if n1N2MsgResp.NgapProcedureCode != ngapType.ProcedureCodeDownlinkNASTransport {
		err = errors.New("unexpected N2 response to N2 UplinkNASTransport")
		return n1N2MsgResp, err
	}
	// Decode NAS PDU
	nasMsg, err := realue.HandleDlInfoTransferEvent(simUe.RealUe, n1N2MsgResp.NasPdu.Value)
	if err != nil {
		return n1N2MsgResp, err
	}
	msgType := nasMsg.GmmHeader.GetMessageType()
	if msgType != nas.MsgTypeDeregistrationAcceptUEOriginatingDeregistration {
		err = errors.New("unexpected N1 response to N1 Registration_Request")
		return n1N2MsgResp, err
	}

	return msgResp, err
}
