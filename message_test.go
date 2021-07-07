package icws_test

import (
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/gildas/go-errors"
	"github.com/gildas/go-icws"
	"github.com/gildas/go-logger"
	"github.com/joho/godotenv"
	"github.com/stretchr/testify/suite"
)

type MessageSuite struct {
	suite.Suite
	Name   string
	Logger *logger.Logger
	Start  time.Time
}

func TestMessageSuite(t *testing.T) {
	suite.Run(t, new(MessageSuite))
}

func (suite *MessageSuite) TestCanUnmarshalUserStatusMessage() {
	payload := suite.LoadTestData("userstatusmessage.json")
	message, err := icws.UnmarshalMessage(payload)
	suite.Require().Nil(err)
	suite.Require().NotNil(message)

	actual, ok := message.(*icws.UserStatusMessage)
	suite.Require().Truef(ok, "Wrong Type: %s", reflect.TypeOf(message).Name())
	suite.Assert().Len(actual.UserStatuses, 4)
}

func (suite *MessageSuite) TestShouldFailUnmarshalWithWrongType() {
	payload := []byte(`{"__type": "boggus", "userStatusList" : []}`)
	_, err := icws.UnmarshalMessage(payload)
	suite.Require().NotNil(err)
	suite.Assert().Truef(errors.Is(err, errors.JSONUnmarshalError), "Error should be a JSONUnmarshalError")
	suite.Assert().Equal(`Unsupported Type "boggus"`, errors.Unwrap(errors.Unwrap(err)).Error())
}

func (suite *MessageSuite) TestShouldFailUnmarshalWithInvalidJSON() {
	payload := []byte(`{'__type': "boggus", "userStatusList" : []}`)
	_, err := icws.UnmarshalMessage(payload)
	suite.Require().NotNil(err)
	suite.Assert().Truef(errors.Is(err, errors.JSONUnmarshalError), "Error should be a JSONUnmarshalError")
	suite.Assert().Equal(`invalid character '\'' looking for beginning of object key string`, errors.Unwrap(errors.Unwrap(err)).Error())
}

// Suite Tools

func (suite *MessageSuite) SetupSuite() {
	_ = godotenv.Load()
	suite.Name = strings.TrimSuffix(reflect.TypeOf(*suite).Name(), "Suite")
	suite.Logger = logger.Create("test",
		&logger.FileStream{
			Path:        fmt.Sprintf("./log/test-%s.log", strings.ToLower(suite.Name)),
			Unbuffered:  true,
			FilterLevel: logger.TRACE,
		},
	).Child("test", "test")
	suite.Logger.Infof("Suite Start: %s %s", suite.Name, strings.Repeat("=", 80-14-len(suite.Name)))
}

func (suite *MessageSuite) TearDownSuite() {
	suite.Logger.Debugf("Tearing down")
	if suite.T().Failed() {
		suite.Logger.Warnf("At least one test failed, we are not cleaning")
		suite.T().Log("At least one test failed, we are not cleaning")
	} else {
		suite.Logger.Infof("All tests succeeded, we are cleaning")
	}
	suite.Logger.Infof("Suite End: %s %s", suite.Name, strings.Repeat("=", 80-12-len(suite.Name)))

	suite.Logger.Infof("Closed the Test WEB Server")
	suite.Logger.Close()
}

func (suite *MessageSuite) BeforeTest(suiteName, testName string) {
	suite.Logger.Infof("Test Start: %s %s", testName, strings.Repeat("-", 80-13-len(testName)))
	suite.Start = time.Now()
}

func (suite *MessageSuite) AfterTest(suiteName, testName string) {
	duration := time.Since(suite.Start)
	if suite.T().Failed() {
		suite.Logger.Errorf("Test %s failed", testName)
	}
	suite.Logger.Record("duration", duration.String()).Infof("Test End: %s %s", testName, strings.Repeat("-", 80-11-len(testName)))
}

func (suite *MessageSuite) LoadTestData(filename string) ([]byte) {
	data, err := os.ReadFile(filepath.Join(".", "testdata", filename))
	if err != nil {
		panic(err)
	}
	return data
}
