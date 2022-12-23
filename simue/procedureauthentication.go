package simue

import (
	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/gnodeb"
	"github.com/omec-project/nas"
	"github.com/omec-project/nas/nasConvert"
	"github.com/omec-project/nas/nasTestpacket"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

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
