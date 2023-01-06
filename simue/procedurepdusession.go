package simue

import (
	"fmt"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/gnodeb/worker/gnbcpueworker"
	"github.com/omec-project/gnbsim/realue"
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

	ProcessN2PduSessionResourceSetupRequest(simUe, n1N2MsgResp.NgapPdu)

	return n1N2MsgResp, err
}

func ProcessN2PduSessionResourceSetupRequest(simUe *simuectx.SimUe,
	ngapPdu *ngapType.NGAPPDU) (err error) {

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
		dst.NASPDU = v.PDUSessionNASPDU
		dst.PDUSessionID = v.PDUSessionID
		dst.SNSSAI = v.SNSSAI
		dst.PDUSessionResourceSetupRequestTransfer = v.PDUSessionResourceSetupRequestTransfer
		list = append(list, dst)
	}

	result, err := gnbcpueworker.ProcessPduSessResourceSetupList(simUe.GnbCpUe, list,
		common.PDU_SESS_RESOURCE_SETUP_REQUEST_EVENT)

	response, err := realue.HandleDataBearerSetupRequestEvent(simUe.RealUe, result)

	err = gnbcpueworker.HandleDataBearerSetupResponse(simUe.GnbCpUe, response)
	return
}
