package shell

import (
	"github.com/kanzihuang/temporal-shell/pkg/common"
	"github.com/stretchr/testify/suite"
	"go.temporal.io/sdk/activity"
	"go.temporal.io/sdk/testsuite"
	"os"
	"path/filepath"
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
	s.env.RegisterActivity(NewActivities(hostTaskQueue))
}

func (s *ActivityTestSuite) begin() common.BeginOutput {
	r := s.Require()
	val, err := s.env.ExecuteActivity(common.Begin)
	r.NoError(err)
	r.True(val.HasValue())
	var output common.BeginOutput
	err = val.Get(&output)
	r.NoError(err)
	r.Equal(hostTaskQueue, output.HostTaskQueue)
	r.Contains(output.SessionDir, hostTaskQueue)
	return output
}

func (s *ActivityTestSuite) end(input common.EndInput) {
	val, err := s.env.ExecuteActivity(common.End, common.EndInput{SessionDir: input.SessionDir})
	s.Require().NoError(err)
	s.Require().True(val.HasValue())
}

func (s *ActivityTestSuite) TestBeginEnd() {
	output := s.begin()
	defer s.end(common.EndInput{SessionDir: output.SessionDir})
}

func (s *ActivityTestSuite) beforeTestReadFile(path string, data []byte) {
	err := os.WriteFile(path, data, 0666)
	s.NoError(err)
}

func (s *ActivityTestSuite) afterTestReadFile(path string) {
	err := os.Remove(path)
	s.NoError(err)
}

func (s *ActivityTestSuite) TestReadFile() {
	tests := []struct {
		name    string
		data    []byte
		wantErr bool
	}{
		{
			name:    "test-read-file-ok",
			data:    []byte("hello world"),
			wantErr: false,
		},
		{
			name:    "test-read-file-valid-size",
			data:    make([]byte, common.BlobSizeMax),
			wantErr: false,
		},
		{
			name:    "test-read-file-too-large",
			data:    make([]byte, common.BlobSizeMax+1),
			wantErr: true,
		},
	}
	beginOutput := s.begin()
	defer s.end(common.EndInput{SessionDir: beginOutput.SessionDir})
	for _, tt := range tests {
		s.Run(tt.name, func() {
			fileName := filepath.Join(beginOutput.SessionDir, tt.name)
			s.beforeTestReadFile(fileName, tt.data)
			defer s.afterTestReadFile(fileName)

			require := s.Require()
			val, err := s.env.ExecuteActivity(common.ReadFile, common.ReadFileInput{
				SessionDir: beginOutput.SessionDir,
				FileName:   tt.name,
			})
			if tt.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			require.True(val.HasValue())

			var output common.ReadFileOutput
			err = val.Get(&output)
			require.NoError(err)
			require.Equal(tt.data, output.Data)
		})
	}
}

func (s *ActivityTestSuite) TestBash() {
	tests := []struct {
		name         string
		command      string
		args         map[string]string
		stdinData    []byte
		wantErr      bool
		wantExitCode int
		withStdout   bool
		wantStdout   []byte
		withStderr   bool
		wantStderr   []byte
	}{
		{
			name:         "command not found",
			command:      "command-not-found",
			wantExitCode: 127,
		},
		{
			name:    "true",
			command: "true",
		},
		{
			name:         "false",
			command:      "false",
			wantExitCode: 1,
		},
		{
			name:    "echo Hello World without stdout",
			command: "echo Hello World",
		},
		{
			name:       "echo Hello World with stdout",
			command:    "echo Hello World",
			withStdout: true,
			wantStdout: []byte("Hello World\n"),
		},
		{
			name:       "echo Hello World then gzip then base64 with stdout",
			command:    "echo Hello World | gzip | base64",
			withStdout: true,
			wantStdout: []byte("H4sIAAAAAAAAA/NIzcnJVwjPL8pJ4QIA4+WVsAwAAAA=\n"),
		}, {
			name:    "echo Hello World without stderr",
			command: "echo Hello World >&2",
		},
		{
			name:       "echo Hello World with stderr",
			command:    "echo Hello World >&2",
			withStderr: true,
			wantStderr: []byte("Hello World\n"),
		},
		{
			name:    "echo arguments with stdout",
			command: "echo I am $name. I am ${age} years old.",
			args: map[string]string{
				"name": "Mike",
				"age":  "18",
			},
			withStdout: true,
			wantStdout: []byte("I am Mike. I am 18 years old.\n"),
		},
		{
			name:       "cat stdio with stdout",
			command:    "cat",
			stdinData:  []byte("Hello World"),
			withStdout: true,
			wantStdout: []byte("Hello World"),
		},
		{
			name:      "cat stdio without stdout",
			command:   "cat",
			stdinData: make([]byte, common.BlobSizeMax+1),
		},
		{
			name:       "cat stdio with too large stdout",
			command:    "cat",
			stdinData:  make([]byte, common.BlobSizeMax+1),
			withStdout: true,
			wantErr:    true,
		},
		{
			name:       "cat stdio with large stderr",
			command:    "cat >&2",
			stdinData:  make([]byte, prefixSuffixLength*2+1),
			withStderr: true,
			wantStderr: append(append(make([]byte, prefixSuffixLength),
				[]byte("\n... omitting 1 bytes ...\n")...), make([]byte, prefixSuffixLength)...),
		},
	}
	beginOutput := s.begin()
	defer s.end(common.EndInput{SessionDir: beginOutput.SessionDir})
	for _, tt := range tests {
		s.Run(tt.name, func() {
			require := s.Require()
			s.env.RegisterActivityWithOptions(BuildBash(tt.command), activity.RegisterOptions{Name: tt.name})
			val, err := s.env.ExecuteActivity(tt.name, common.BashInput{
				WithStdout: tt.withStdout,
				WithStderr: tt.withStderr,
				Args:       tt.args,
				StdinData:  tt.stdinData,
			})
			if tt.wantErr {
				require.Error(err)
				return
			}
			require.NoError(err)
			require.True(val.HasValue())

			var output common.BashOutput
			err = val.Get(&output)
			require.NoError(err)
			require.Equal(tt.wantExitCode, output.ExitCode)
			require.Equal(tt.wantStdout, output.StdoutData)
			require.Equal(tt.wantStderr, output.StderrData)
		})
	}
}
