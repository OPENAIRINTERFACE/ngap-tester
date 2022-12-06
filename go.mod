module github.com/openairinterface/ngap-tester

go 1.15

replace (
	github.com/omec-project/gnbsim => ./third-party/gnbsim
	github.com/omec-project/gnbsim/logger => ./third-party/gnbsim/logger
	github.com/omec-project/nas => ./third-party/nas
	github.com/omec-project/ngap => ./third-party/ngap
)

require (
	github.com/omec-project/gnbsim v0.0.0-00010101000000-000000000000
	github.com/sirupsen/logrus v1.8.1
	github.com/urfave/cli v1.22.4
)
