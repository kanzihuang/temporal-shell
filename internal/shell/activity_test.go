package shell

import (
	"github.com/kanzihuang/temporal-shell/pkg/common"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"testing"
)

const hostTaskQueue = "testHostTaskQueue"

func TestActivityTestSuite(t *testing.T) {
	suite.Run(t, new(ActivityTestSuite))
}

type ActivityTestSuite struct {
	suite.Suite
	testsuite.WorkflowTestSuite
	env *testsuite.TestActivityEnvironment
}

func (s *ActivityTestSuite) SetupTest() {
	s.env = s.NewTestActivityEnvironment()
	s.env.RegisterActivityWithOptions(BuildGetHostTaskQueue(hostTaskQueue), activity.RegisterOptions{Name: common.GetHostTaskQueue})
	s.env.RegisterActivityWithOptions(ReadFile, activity.RegisterOptions{Name: common.Download})
}

func (s *ActivityTestSuite) TestGetHostTaskQueue() {
	val, err := s.env.ExecuteActivity(common.GetHostTaskQueue)
	s.NoError(err)
	s.True(val.HasValue())

	var taskQueue string
	err = val.Get(&taskQueue)
	s.NoError(err)
	s.Equal(hostTaskQueue, taskQueue)
}

func (s *ActivityTestSuite) TestDownload() {
	val, err := s.env.ExecuteActivity(common.Download)
	s.NoError(err)
	s.True(val.HasValue())

	var taskQueue string
	err = val.Get(&taskQueue)
	s.NoError(err)
	s.Equal(hostTaskQueue, taskQueue)
}
