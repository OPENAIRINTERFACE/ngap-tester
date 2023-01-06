// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package simue

import (
	"fmt"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/gnodeb"
	realue_nas "github.com/omec-project/gnbsim/realue/nas"
	"github.com/omec-project/nas"
	"github.com/omec-project/nas/nasMessage"
	"github.com/omec-project/nas/nasTestpacket"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

func SendRegisterRequest(simUe *simuectx.SimUe) error {
	var err error
	//-------------------
	// Step send NAS Registration_Request
	//-------------------
	nasPdu, err := realue_nas.GetRegisterRequest(simUe.RealUe)
	if err != nil {
		return err
	}
	sendMsg, err := gnodeb.GetInitialUEMessage(simUe.GnB, simUe.GnbCpUe, nasPdu)
	if err != nil {
		return err
	}
	msg := FormN2Message(common.N2_SEND_SDU_EVENT, sendMsg)
	SendToGnbUe(simUe, msg)
	return nil
}

func SendDeregisterRequest(simUe *simuectx.SimUe) error {
	var err error
	nasPdu, err := realue_nas.GetDeregisterRequest(simUe.RealUe)
	if err != nil {
		return err
	}

	sendMsg, err := gnodeb.GetUplinkNASTransport(simUe.GnB, simUe.GnbCpUe, nasPdu)
	if err != nil {
		return err
	}
	msg := FormN2Message(common.N2_SEND_SDU_EVENT, sendMsg)
	SendToGnbUe(simUe, msg)
	return nil
}

func SendInitialContextSetupResponse(simUe *simuectx.SimUe) error {
	var err error
	sendMsg, err := gnodeb.GetInitialContextSetupResponse(simUe.GnB, simUe.GnbCpUe)
	if err != nil {
		return err
	}
	msg := FormN2Message(common.N2_SEND_SDU_EVENT, sendMsg)
	SendToGnbUe(simUe, msg)
	return nil
}

func SendRegistrationComplete(simUe *simuectx.SimUe) error {
	var err error
	nasPdu, err := realue_nas.GetRegistrationComplete(simUe.RealUe)
	if err != nil {
		simUe.Log.Errorln("Failed to encode NAS-Registration_Complete NAS Message due to", err)
		return err
	}
	sendMsg, err := gnodeb.GetUplinkNASTransport(simUe.GnB, simUe.GnbCpUe, nasPdu)
	if err != nil {
		return err
	}
	msg := FormN2Message(common.N2_SEND_SDU_EVENT, sendMsg)
	SendToGnbUe(simUe, msg)
	return nil
}

func SendPduSessionEstablishmentRequest(simUe *simuectx.SimUe) error {

	nasPdu := nasTestpacket.GetUlNasTransport_PduSessionEstablishmentRequest(10,
		nasMessage.ULNASTransportRequestTypeInitialRequest, simUe.RealUe.Dnn, simUe.RealUe.SNssai)

	nasPdu, err := realue_nas.EncodeNasPduWithSecurity(simUe.RealUe, nasPdu,
		nas.SecurityHeaderTypeIntegrityProtectedAndCiphered, true)
	if err != nil {
		fmt.Println("Failed to encrypt PDU Session Establishment Request Message", err)
		return err
	}
	sendMsg, err := gnodeb.GetUplinkNASTransport(simUe.GnB, simUe.GnbCpUe, nasPdu)
	if err != nil {
		return err
	}
	msg := FormN2Message(common.N2_SEND_SDU_EVENT, sendMsg)
	SendToGnbUe(simUe, msg)
	return nil
}
