package simue

import (
	"errors"
	"fmt"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/gnodeb"
	"github.com/omec-project/gnbsim/realue"
	realue_nas "github.com/omec-project/gnbsim/realue/nas"
	"github.com/omec-project/nas"
	"github.com/omec-project/nas/nasConvert"
	"github.com/omec-project/nas/nasMessage"
	"github.com/omec-project/nas/nasTestpacket"
	"github.com/omec-project/nas/nasType"
	"github.com/omec-project/ngap/ngapType"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

func PerformRegisterProcedure(simUe *simuectx.SimUe) (common.InterfaceMessage, error) {
	var err error
	simUe.Log.Traceln("PerformRegisterProcedure")
	//-------------------
	// Step send NAS Registration_Request
	//-------------------
	nasPdu, err := realue_nas.GetRegisterRequest(simUe.RealUe)
	if err != nil {
		return nil, err
	}
	sendMsg, err := gnodeb.GetInitialUEMessage(simUe.GnB, simUe.GnbCpUe, nasPdu)
	if err != nil {
		return nil, err
	}
	msg := FormN2Message(common.N2_SEND_SDU_EVENT, sendMsg)
	SendToGnbUe(simUe, msg)

	//-------------------
	// Step expect receive NGAP-DownlinkNASTransport/NAS-Authentication_Request
	// LG: TODO a func dedicated to this ?
	//-------------------
	msgResp, ok := simUe.RcvTimedSecondEvent(3)
	if !ok {
		err = errors.New("expected NGAP-DownlinkNASTransport/NAS-Authentication_Request coming from AMF timed-out")
		return nil, err
	}
	if msgResp.GetEventType() != common.N1_N2_RECV_SDU_EVENT {
		err = errors.New("unexpected response to N1 Registration-Request")
		return msgResp, err
	}
	n1N2MsgResp := msgResp.(*common.N1N2Message)
	if n1N2MsgResp.NgapProcedureCode != ngapType.ProcedureCodeDownlinkNASTransport {
		err = errors.New("unexpected N2 response to N2 InitialUEMessage")
		return n1N2MsgResp, err
	}
	// Decode NAS PDU
	nasMsg, err := realue.HandleDlInfoTransferEvent(simUe.RealUe, n1N2MsgResp.NasPdu.Value)
	if err != nil {
		return n1N2MsgResp, err
	}
	msgType := nasMsg.GmmHeader.GetMessageType()
	if msgType != nas.MsgTypeAuthenticationRequest {
		err = errors.New("unexpected N1 response to N1 Registration_Request")
		return n1N2MsgResp, err
	}

	//-------------------
	// Step answer to NGAP-DownlinkNASTransport/NAS-Authentication_Request
	//-------------------
	err = PerformAuthenticationProcedure(simUe, nasMsg)
	if err != nil {
		return n1N2MsgResp, err
	}
	//-------------------
	// Step expect receive  NGAP-DownlinkNASTransport/NAS-Security_Mode_Command
	// LG: TODO a func dedicated to this
	//-------------------
	msgResp, ok = simUe.RcvTimedSecondEvent(3)
	if !ok {
		err = errors.New("expected NGAP-DownlinkNASTransport/NAS-Security_Mode_Command coming from AMF timed-out")
		return nil, err
	}

	if msgResp.GetEventType() != common.N1_N2_RECV_SDU_EVENT {
		err = errors.New("expected message to N1 Authentication_Response")
		return msgResp, err
	}
	n1N2MsgResp = msgResp.(*common.N1N2Message)
	if n1N2MsgResp.NgapProcedureCode != ngapType.ProcedureCodeDownlinkNASTransport {
		err = errors.New("unexpected N2 message, expected DownlinkNASTransport")
		return n1N2MsgResp, err
	}
	// Decode NAS PDU
	nasMsg, err = realue.HandleDlInfoTransferEvent(simUe.RealUe, n1N2MsgResp.NasPdu.Value)
	if err != nil {
		return n1N2MsgResp, err
	}
	msgType = nasMsg.GmmHeader.GetMessageType()
	if msgType != nas.MsgTypeSecurityModeCommand {
		err = errors.New("unexpected N1 message, expected Security_Mode_Command")
		return n1N2MsgResp, err
	}
	//-------------------
	// Step answer to NGAP-DownlinkNASTransport/NAS-Security_Mode_Command
	//-------------------
	err = PerformSecurityProcedure(simUe, nasMsg)
	if err != nil {
		return n1N2MsgResp, err
	}

	//-------------------
	// Step expect receive  NGAP-InitialContextSetup/NAS-Register_Accept
	// LG: TODO a func dedicated to this
	//-------------------
	msgResp, ok = simUe.RcvTimedSecondEvent(3)
	if !ok {
		err = errors.New("expected NGAP-InitialContextSetupRequest/NAS-Register_Accept coming from AMF timed-out")
		return nil, err
	}

	if msgResp.GetEventType() != common.N1_N2_RECV_SDU_EVENT {
		err = errors.New("expected message to N1 Authentication_Response")
		return msgResp, err
	}
	n1N2MsgResp = msgResp.(*common.N1N2Message)
	if n1N2MsgResp.NgapProcedureCode != ngapType.ProcedureCodeInitialContextSetup {
		err = errors.New("unexpected N2 message, expected InitialContextSetupRequest")
		return n1N2MsgResp, err
	}

	if n1N2MsgResp.NasPdu == nil {
		err = errors.New("missing expected NAS-Register_Accept")
		return n1N2MsgResp, err
	}
	nasMsg, err = realue_nas.NASDecode(simUe.RealUe, nas.GetSecurityHeaderType(n1N2MsgResp.NasPdu.Value), n1N2MsgResp.NasPdu.Value)
	if err != nil {
		simUe.Log.Errorln("Failed to decode NAS Message due to", err)
		return n1N2MsgResp, err
	}
	// Process IEs
	err = realue.HandleRegistrationAccept(simUe.RealUe, nasMsg.RegistrationAccept)
	if err != nil {
		return nil, err
	}
	//-------------------
	// Step Send back  NGAP-InitialContextSetupResponse/NAS-none
	//-------------------
	sendMsg, err = gnodeb.GetInitialContextSetupResponse(simUe.GnB, simUe.GnbCpUe)
	if err != nil {
		return nil, err
	}
	msg = FormN2Message(common.N2_SEND_SDU_EVENT, sendMsg)
	SendToGnbUe(simUe, msg)
	//-------------------
	// Step Send back  NGAP-DownlinkNASTransport/NAS-Registration_Complete
	//-------------------

	nasPdu, err = realue_nas.GetRegistrationComplete(simUe.RealUe)
	if err != nil {
		simUe.Log.Errorln("Failed to encode NAS-Registration_Complete NAS Message due to", err)
		return nil, err
	}
	sendMsg, err = gnodeb.GetUplinkNASTransport(simUe.GnB, simUe.GnbCpUe, nasPdu)
	if err != nil {
		return nil, err
	}
	msg = FormN2Message(common.N2_SEND_SDU_EVENT, sendMsg)
	SendToGnbUe(simUe, msg)

	return msgResp, err

}

func PerformAuthenticationProcedure(simUe *simuectx.SimUe, nasMsg *nas.Message) error {
	simUe.Log.Traceln("PerformAuthenticationProcedure")
	authReq := nasMsg.AuthenticationRequest

	simUe.RealUe.NgKsi = nasConvert.SpareHalfOctetAndNgksiToModels(authReq.SpareHalfOctetAndNgksi)

	rand := authReq.GetRANDValue()
	autn := authReq.GetAUTN()
	mnc := simUe.RealUe.Plmn.Mnc
	if len(mnc) == 2 {
		mnc = "0" + mnc
	}
	snName := "5G:mnc" + mnc + ".mcc" + simUe.RealUe.Plmn.Mcc + ".3gppnetwork.org"

	resStat := simUe.RealUe.DeriveRESstarAndSetKey(autn[:], rand[:], snName)

	// TODO: Parse Auth Request IEs and update the RealUE Context

	// Now generate NAS Authentication Response
	nasPdu := nasTestpacket.GetAuthenticationResponse(resStat, "")

	sendMsg, err := gnodeb.GetUplinkNASTransport(simUe.GnB, simUe.GnbCpUe, nasPdu)
	if err != nil {
		return err
	}
	msg := FormN2Message(common.N2_SEND_SDU_EVENT, sendMsg)
	SendToGnbUe(simUe, msg)
	return err
}

func PerformSecurityProcedure(simUe *simuectx.SimUe, nasMsg *nas.Message) (err error) {
	simUe.Log.Traceln("PerformSecurityProcedure")
	//TODO: Process corresponding Security Mode Command first

	mobileId5GS := nasType.MobileIdentity5GS{
		Len:    uint16(len(simUe.RealUe.Suci)), // suci
		Buffer: simUe.RealUe.Suci,
	}
	registrationRequestWith5GMM := nasTestpacket.GetRegistrationRequest(
		nasMessage.RegistrationType5GSInitialRegistration, mobileId5GS, nil,
		simUe.RealUe.GetUESecurityCapability(), simUe.RealUe.Get5GMMCapability(), nil, nil)

	simUe.Log.Traceln("Generating Security Mode Complete Message")
	nasPdu := nasTestpacket.GetSecurityModeComplete(registrationRequestWith5GMM)

	nasPdu, err = realue_nas.EncodeNasPduWithSecurity(simUe.RealUe, nasPdu,
		nas.SecurityHeaderTypeIntegrityProtectedAndCipheredWithNew5gNasSecurityContext,
		true)
	if err != nil {
		simUe.Log.Errorln("EncodeNasPduWithSecurity() returned:", err)
		return fmt.Errorf("failed to encrypt security mode complete message")
	}

	sendMsg, err := gnodeb.GetUplinkNASTransport(simUe.GnB, simUe.GnbCpUe, nasPdu)
	if err != nil {
		return err
	}
	msg := FormN2Message(common.N2_SEND_SDU_EVENT, sendMsg)
	SendToGnbUe(simUe, msg)
	return err
}
