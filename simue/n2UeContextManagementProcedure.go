package simue

import (
	"fmt"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/util/test"
	"github.com/omec-project/ngap/ngapType"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

func PerformUEContextReleaseProcedureNwInitiated(
	simUe *simuectx.SimUe,
	pdu *ngapType.NGAPPDU,
	assertedCausePresent int,
	assertedCauseValue int,
	assertCause bool,
) (err error) {
	var ueNgapIds *ngapType.UENGAPIDs
	var amfUeNgapId ngapType.AMFUENGAPID
	var cause *ngapType.Cause

	simUe.Log.Traceln("PerformUEContextReleaseProcedureNwInitiated")

	initiatingMessage := pdu.InitiatingMessage
	ueCtxRelCmd := initiatingMessage.Value.UEContextReleaseCommand

	for _, ie := range ueCtxRelCmd.ProtocolIEs.List {
		switch ie.Id.Value {
		case ngapType.ProtocolIEIDUENGAPIDs:
			ueNgapIds = ie.Value.UENGAPIDs
			if ueNgapIds == nil {
				err = fmt.Errorf("UENGAPIDs is nil")
				return
			}
		case ngapType.ProtocolIEIDCause:
			cause = ie.Value.Cause
			if cause == nil {
				err = fmt.Errorf("Mandatory IE Cause is nil")
				return
			}

			if assertCause {
				if !test.EqualCause(
					cause,
					ngapType.CausePresentMisc,
					ngapType.CauseMiscPresentUnspecified,
				) {
					err = fmt.Errorf("Assert bad IE Cause")
					return
				}
			}
		}
	}

	if ueNgapIds.Present == ngapType.UENGAPIDsPresentUENGAPIDPair {
		amfUeNgapId = ueNgapIds.UENGAPIDPair.AMFUENGAPID
		if simUe.GnbCpUe.AmfUeNgapId != amfUeNgapId.Value {
			err = fmt.Errorf("AmfUeNgapId mismatch")
			return
		}
	}

	var pduSessIds []int64
	f := func(k interface{}, v interface{}) bool {
		pduSessIds = append(pduSessIds, k.(int64))
		return true
	}
	simUe.GnbCpUe.GnbUpUes.Range(f)

	ngapPdu, err := test.GetUEContextReleaseComplete(simUe.GnbCpUe.AmfUeNgapId,
		simUe.GnbCpUe.GnbUeNgapId, pduSessIds)
	if err != nil {
		err = fmt.Errorf("Failed to create UE Context Release Complete message")
		return
	}

	err = simUe.GnbCpUe.Gnb.CpTransport.SendToPeer(simUe.GnbCpUe.Amf, ngapPdu)
	if err != nil {
		simUe.Log.Errorln("SendToPeer failed:", err)
		err = fmt.Errorf(
			"SendToPeer failed to send UE Context Release Complete message",
		)
		return
	}
	simUe.Log.Traceln("Sent UE Context Release Complete Message to AMF")

	quitEvt := &common.DefaultMessage{}
	quitEvt.Event = common.QUIT_EVENT
	simUe.GnbCpUe.ReadChan <- quitEvt

	return err
}
