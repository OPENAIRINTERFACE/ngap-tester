// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package simue

import (
	"fmt"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/gnodeb"
	gnbctx "github.com/omec-project/gnbsim/gnodeb/context"
	"github.com/omec-project/gnbsim/realue"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

func InitUE(imsiStr string, ueModel string, gnb *gnbctx.GNodeB) *simuectx.SimUe {
	simUe := simuectx.NewSimUe(imsiStr, ueModel, gnb)
	Init(simUe) // Initialize simUE, realUE & wait for events
	return simUe
}

func Init(simUe *simuectx.SimUe) error {

	err := ConnectToGnb(simUe)
	if err != nil {
		err = fmt.Errorf("Failed to connect to gnodeb: %v!", err)
		return err
	}

	err = realue.Init(simUe.RealUe)
	return err
}

func ConnectToGnb(simUe *simuectx.SimUe) error {
	uemsg := common.UuMessage{}
	uemsg.Event = common.CONNECTION_REQUEST_EVENT
	uemsg.CommChan = simUe.ReadChan
	uemsg.Supi = simUe.Supi

	var err error
	gNb := simUe.GnB
	simUe.WriteGnbUeChan, simUe.GnbCpUe, err = gnodeb.RequestConnection(gNb, &uemsg)
	if err != nil {
		simUe.Log.Infof("ERROR -- connecting to gNodeB, Name:%v, IP:%v, Port:%v", gNb.GnbName,
			gNb.GnbN2Ip, gNb.GnbN2Port)
		return err
	}

	simUe.Log.Infof("Connected to gNodeB, Name:%v, IP:%v, Port:%v", gNb.GnbName,
		gNb.GnbN2Ip, gNb.GnbN2Port)
	return nil
}

func SendToGnbUe(simUe *simuectx.SimUe, msg common.InterfaceMessage) {
	simUe.Log.Traceln("Sending", msg.GetEventType(), "to GnbUe")
	simUe.WriteGnbUeChan <- msg
}

func FormN2Message(event common.EventType, n2Pdu []byte) *common.N2EncodedMessage {
	msg := &common.N2EncodedMessage{}
	msg.Event = event
	msg.N2Pdus = n2Pdu
	return msg
}
