package simue

import (
	"fmt"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/gnodeb"
	realue_nas "github.com/omec-project/gnbsim/realue/nas"
	"github.com/omec-project/nas"
	"github.com/omec-project/nas/nasMessage"
	"github.com/omec-project/nas/nasTestpacket"
	"github.com/omec-project/nas/nasType"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

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
