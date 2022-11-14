package cmd

import (
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/suite"
	"k8s.io/apimachinery/pkg/version"
)

var FaikedVersion = &version.Info{
	Major: "1",
	Minor: "23",
}

type TestSuite struct {
	suite.Suite
}

func (s *TestSuite) SetupSuite() {

	// Init logger
	logrus.SetLevel(logrus.DebugLevel)

}

func TestTestSuite(t *testing.T) {
	suite.Run(t, new(TestSuite))
}
