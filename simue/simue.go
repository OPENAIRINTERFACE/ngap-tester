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
	realue_nas "github.com/omec-project/gnbsim/realue/nas"
	"github.com/omec-project/nas"
	simuectx "github.com/openairinterface/ngap-tester/simue/context"
)

func InitUE(
	imsiStr string,
	ueModel string,
	gnb *gnbctx.GNodeB,
) *simuectx.SimUe {
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
	simUe.WriteGnbUeChan, simUe.GnbCpUe, err = gnodeb.RequestConnection(
		gNb,
		&uemsg,
	)
	if err != nil {
		simUe.Log.Infof(
			"ERROR -- connecting to gNodeB, Name:%v, IP:%v, Port:%v",
			gNb.GnbName,
			gNb.GnbN2Ip,
			gNb.GnbN2Port,
		)
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

func FormN2Message(
	event common.EventType,
	n2Pdu []byte,
) *common.N2EncodedMessage {
	msg := &common.N2EncodedMessage{}
	msg.Event = event
	msg.N2Pdus = n2Pdu
	return msg
}

func ExpectReceiveN1N2(
	simUe *simuectx.SimUe,
	ngapProcedureCode int64,
	nasMsgType uint8,
	timeOutSeconds int,
) (n1N2MsgResp *common.N1N2Message, n1DecodedMsg *nas.Message, err error) {
	msgResp, ok := simUe.RcvTimedSecondEvent(timeOutSeconds)
	if !ok {
		err = fmt.Errorf(
			"expected NGAP %v / NAS %v: coming from AMF timed-out",
			err,
		)
		return
	}
	if msgResp.GetEventType() != common.N1_N2_RECV_SDU_EVENT {
		err = fmt.Errorf(
			"unexpected event %v, wanted: N1_N2_RECV_SDU_EVENT",
			msgResp.GetEventType().String(),
		)
		return
	}
	n1N2MsgResp = msgResp.(*common.N1N2Message)
	if n1N2MsgResp.NgapProcedureCode != ngapProcedureCode {
		err = fmt.Errorf(
			"unexpected NGAP procedure code %v, wanted: %v",
			n1N2MsgResp.NgapProcedureCode,
			ngapProcedureCode,
		)
		return
	}
	if nasMsgType != 0 {
		if n1N2MsgResp.NasPdu == nil {
			err = fmt.Errorf("NAS message (type %v) missing", nasMsgType)
			return
		}
		n1DecodedMsg, err = realue_nas.NASDecode(
			simUe.RealUe,
			nas.GetSecurityHeaderType(n1N2MsgResp.NasPdu.Value),
			n1N2MsgResp.NasPdu.Value,
		)
		msgType := n1DecodedMsg.GmmHeader.GetMessageType()
		if msgType != nasMsgType {
			if msgType == nas.MsgTypeDLNASTransport {
				payload := n1DecodedMsg.GmmMessage.DLNASTransport.PayloadContainer
				if payload.Len != 0 {
					buffer := payload.Buffer[:payload.Len]
					m := nas.NewMessage()
					err = m.PlainNasDecode(&buffer)
					if err != nil {
						err = fmt.Errorf("failed to decode payload container")
						return
					}
					msgType = m.GsmHeader.GetMessageType()
					if msgType != nasMsgType {
						err = fmt.Errorf(
							"NAS message (type %v) unexpected (type %v)",
							msgType,
							nasMsgType,
						)
						return
					}
					// remove wrapper of container
					n1DecodedMsg = m
				}
			} else {
				err = fmt.Errorf("NAS message (type %v) unexpected (type %v)", msgType, nasMsgType)
				return
			}
		}
	} else {
		if n1N2MsgResp.NasPdu != nil {
			n1DecodedMsg, err = realue_nas.NASDecode(simUe.RealUe, nas.GetSecurityHeaderType(n1N2MsgResp.NasPdu.Value), n1N2MsgResp.NasPdu.Value)
			if err != nil {
				msgType := n1DecodedMsg.GmmHeader.GetMessageType()
				err = fmt.Errorf("NAS message unexpected (type %v)", msgType)
				return
			}
		}
	}
	simUe.Log.Traceln("Expected event occured N2 proc code ", ngapProcedureCode,
		" N1 msg type ", nasMsgType)
	return
}
