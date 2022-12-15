// SPDX-FileCopyrightText: 2021 Open Networking Foundation <info@opennetworking.org>
//
// SPDX-License-Identifier: Apache-2.0

package context

import (
	"sync"
	"time"

	"github.com/omec-project/gnbsim/common"
	"github.com/omec-project/gnbsim/factory"
	gnbctx "github.com/omec-project/gnbsim/gnodeb/context"
	"github.com/omec-project/gnbsim/logger"
	realuectx "github.com/omec-project/gnbsim/realue/context"
	"github.com/omec-project/nas/security"

	"github.com/sirupsen/logrus"
)

func init() {
	SimUeTable = make(map[string]*SimUe)
}

type LastIndexesProvision struct {
	Ausf int `json:"ausf"`
}
type SequenceNumberProvision struct {
	Sqn         string               `json:"sqn"`
	SqnScheme   string               `json:"sqnScheme"`
	LastIndexes LastIndexesProvision `json:"protectionParameterId"`
}

type SubscriberProvision struct {
	AuthenticationMethod          string                  `json:"authenticationMethod"`
	EncPermanentKey               string                  `json:"encPermanentKey"`
	ProtectionParameterId         string                  `json:"protectionParameterId"`
	SequenceNumber                SequenceNumberProvision `json:"sequenceNumber"`
	AuthenticationManagementField string                  `json:"authenticationManagementField"`
	AlgorithmId                   string                  `json:"algorithmId"`
	EncOpcKey                     string                  `json:"encOpcKey"`
	EncTopcKey                    string                  `json:"encTopcKey"`
	VectorGenerationInHss         bool                    `json:"vectorGenerationInHss"`
	N5gcAuthMethod                string                  `json:"n5gcAuthMethod"`
	RgAuthenticationInd           bool                    `json:"rgAuthenticationInd"`
	Supi                          string                  `json:"supi"`
}

type SimUe struct {
	Supi      string
	GnB       *gnbctx.GNodeB
	RealUe    *realuectx.RealUe
	Procedure common.ProcedureType
	WaitGrp   sync.WaitGroup

	// SimUe writes messages to GnbUE on this channel
	WriteGnbUeChan chan common.InterfaceMessage

	// whatever to UE scenario
	ReadChan chan common.InterfaceMessage

	/* logger */
	Log *logrus.Entry
}

var SimUeTable map[string]*SimUe

func NewSimUe(supi string, ueModel string, gnb *gnbctx.GNodeB) *SimUe {
	ueProfile, err := factory.AppConfig.Configuration.GetUeProfile(ueModel)
	if err != nil {
		return nil
	}
	simue := SimUe{}
	simue.GnB = gnb
	simue.Supi = supi
	simue.ReadChan = make(chan common.InterfaceMessage, 5)
	// TODO select prefered security algorithms
	simue.RealUe = realuectx.NewRealUe(supi,
		security.AlgCiphering128NEA0, security.AlgIntegrity128NIA2,
		simue.ReadChan, ueProfile.Plmn, ueProfile.Key, ueProfile.Opc, ueProfile.Nas.SeqNum,
		ueProfile.Nas.Dnn, ueProfile.Nas.SNssai)

	simue.Log = logger.SimUeLog.WithField(logger.FieldSupi, supi)

	simue.Log.Traceln("Created new SimUe context")
	SimUeTable[supi] = &simue
	return &simue
}

func GetSimUe(supi string) *SimUe {
	simue, found := SimUeTable[supi]
	if found == false {
		return nil
	}
	return simue
}

func (ue *SimUe) RcvTimedMilliSecondEvent(timeOutMilliSeconds int) (common.InterfaceMessage, bool) {
	select {
	case msg, ok := <-ue.ReadChan:
		ue.Log.Traceln("Received event ", msg.GetEventType())
		return msg, ok

	case <-time.After(time.Duration(timeOutMilliSeconds) * time.Millisecond):
		return nil, false
	}
}

func (ue *SimUe) RcvTimedSecondEvent(timeOutSeconds int) (common.InterfaceMessage, bool) {
	select {
	case msg, ok := <-ue.ReadChan:
		ue.Log.Traceln("Received event ", msg.GetEventType())
		return msg, ok

	case <-time.After(time.Duration(timeOutSeconds) * time.Second):
		return nil, false
	}
}

func (ue *SimUe) RcvEvent() (common.InterfaceMessage, bool) {
	msg, ok := <-ue.ReadChan
	ue.Log.Traceln("Received event ", msg.GetEventType())
	return msg, ok
}
