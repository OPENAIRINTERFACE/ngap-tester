package simue

import (
	"errors"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/gnodeb"
	"github.com/omec-project/gnbsim/realue"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

func PerformRegisterProcedure(simUe *simuectx.SimUe) (common.InterfaceMessage, error) {
	var err error

	nasPdu, err := realue.EncodeRegRequestEvent(simUe.RealUe)
	if err != nil {
		return nil, err
	}
	sendMsg, err := gnodeb.GetInitialUEMessage(simUe.GnB, simUe.GnbCpUe, nasPdu)

	msg := FormN2Message(common.N2_SEND_SDU_EVENT, sendMsg)
	SendToGnbUe(simUe, msg)

	msgResp, ok := simUe.RcvTimedSecondEvent(3)
	if !ok {
		err = errors.New("response to N1 Registration-Request timed-out")
	}

	return msgResp, err

}
