package simue

import (
	"fmt"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/gnodeb/worker/gnbcpueworker"
	"github.com/omec-project/gnbsim/realue"
	realuenas "github.com/omec-project/gnbsim/realue/nas"
	"github.com/omec-project/nas"
	"github.com/omec-project/ngap/ngapType"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

func PerformPduSessionEstablishmentProcedure(simUe *simuectx.SimUe) (common.InterfaceMessage, error) {
	var err error
	simUe.Log.Traceln("PerformDeregisterProcedureUEOriginatingDeregistration")
	//-------------------
	// Step send NAS PDU_Session_establishment_Request
	//-------------------
	err = SendPduSessionEstablishmentRequest(simUe)
	if err != nil {
		return nil, err
	}

	//-------------------
	// Step expect receive NGAP-PDUSessionResourceSetup/NAS-DLNASTransport/
	// NAS-PDU_Session_establishment_Accept
	n1N2MsgResp, _, err := ExpectReceiveN1N2(simUe,
		ngapType.ProcedureCodePDUSessionResourceSetup,
		0, 3) // TODO how to assert NAS msg list ?

	pduSessions, err := ProcessN2PduSessionResourceSetupRequest(simUe, n1N2MsgResp.NgapPdu)
	if err != nil {
		return nil, err
	}

	err = SendPduSessionResourceSetupResponse(simUe, pduSessions)
	if err != nil {
		return nil, err
	}

	return n1N2MsgResp, err
}

func PduSessionGenerateULTraffic(simUe *simuectx.SimUe) (err error) {
	msg := &common.UeMessage{}
	// Actually can do only one
	msg.UserDataPktCount = 1
	// Actually unused
	msg.UserDataPktInterval = 1
	// LG TODO hardcoded
	msg.DefaultAs = "192.168.0.254"
	msg.Event = common.DATA_PKT_GEN_REQUEST_EVENT
	err = realue.HandleDataPktGenRequestEvent(simUe.RealUe, msg)
	return
}

func ProcessN2PduSessionResourceSetupRequest(simUe *simuectx.SimUe, ngapPdu *ngapType.NGAPPDU) (response *common.UuMessage, err error) {

	var amfUeNgapId *ngapType.AMFUENGAPID
	var pduSessResourceSetupReqList *ngapType.PDUSessionResourceSetupListSUReq

	initiatingMessage := ngapPdu.InitiatingMessage
	pduSessResourceSetupReq := initiatingMessage.Value.PDUSessionResourceSetupRequest

	for _, ie := range pduSessResourceSetupReq.ProtocolIEs.List {
		switch ie.Id.Value {
		case ngapType.ProtocolIEIDAMFUENGAPID:
			amfUeNgapId = ie.Value.AMFUENGAPID
			if amfUeNgapId == nil {
				err = fmt.Errorf("AMFUENGAPID is nil")
				return
			}
		case ngapType.ProtocolIEIDPDUSessionResourceSetupListSUReq:
			pduSessResourceSetupReqList = ie.Value.PDUSessionResourceSetupListSUReq
			if pduSessResourceSetupReqList == nil || len(pduSessResourceSetupReqList.List) == 0 {
				err = fmt.Errorf("PDUSessionResourceSetupListSUReq is empty")
				return
			}
		}
	}

	var list []gnbcpueworker.PduSessResourceSetupItem
	for _, v := range pduSessResourceSetupReqList.List {
		dst := gnbcpueworker.PduSessResourceSetupItem{}
		if v.PDUSessionNASPDU != nil {
			pkg := []byte(v.PDUSessionNASPDU.Value)
			m, er := realuenas.NASDecode(simUe.RealUe, nas.GetSecurityHeaderType(pkg), pkg)
			if er != nil {
				err = er
				return
			}
			msgType := m.GmmHeader.GetMessageType()
			if msgType != nas.MsgTypeDLNASTransport {
				err = fmt.Errorf("not a 5GMM DLNASTransport message")
				return
			}
			nasMsgType, nasMsg, er := realuenas.NasGetTransferContent(simUe.RealUe, m)

			if nasMsgType == nas.MsgTypePDUSessionEstablishmentAccept {
				msg := &common.UeMessage{}
				msg.NasMsg = nasMsg
				realue.HandlePduSessEstAcceptEvent(simUe.RealUe, msg)
			} else {
				err = fmt.Errorf("not a 5GMM DLNASTransport message")
				return
			}
		}
		dst.NASPDU = v.PDUSessionNASPDU
		dst.PDUSessionID = v.PDUSessionID
		dst.SNSSAI = v.SNSSAI
		dst.PDUSessionResourceSetupRequestTransfer = v.PDUSessionResourceSetupRequestTransfer
		list = append(list, dst)
	}

	result, err := gnbcpueworker.ProcessPduSessResourceSetupList(simUe.GnbCpUe, list,
		common.PDU_SESS_RESOURCE_SETUP_REQUEST_EVENT)
	if err != nil {
		return
	}
	response, err = realue.HandleDataBearerSetupRequestEvent(simUe.RealUe, result)
	if err != nil {
		return
	}
	err = gnbcpueworker.HandleDataBearerSetupResponse(simUe.GnbCpUe, response)
	return
}
