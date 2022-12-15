package simue

import (
	"errors"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/realue"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

func PerformRegisterProcedure(simUe *simuectx.SimUe) (common.InterfaceMessage, error) {
	var err error

	nasPdu, err := realue.EncodeRegRequestEvent(simUe.RealUe)
	if err != nil {
		return nil, err
	}
	msg := realue.FormUuMessage(common.N1_SEND_SDU_EVENT+common.NAS_5GMM_REGISTRATION_REQUEST, nasPdu)
	msg.Tac = simUe.GnB.SupportedTaList[0].Tac
	msg.NrCgi = simUe.GnB.NrCgiCellList[0]
	SendToGnbUe(simUe, msg)

	msgResp, ok := simUe.RcvTimedSecondEvent(3)
	if !ok {
		err = errors.New("Response to N1 Registration-Request timed-out !")
	}
	return msgResp, err

}
