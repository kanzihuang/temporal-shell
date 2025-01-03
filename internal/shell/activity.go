package shell

import (
	"context"
	"github.com/kanzihuang/temporal-shell/pkg/common"
)

func ReadFile(ctx context.Context, input common.ReadFileInput) (common.ReadFileOutput, error) {
	panic("implement me")
}

func BuildGetHostTaskQueue(hostTaskQueue string) func() (string, error) {
	return func() (string, error) {
		return hostTaskQueue, nil
	}
}

func BuildExecute(command string) func(ctx context.Context, input common.ActivityInput) (common.ActivityOutput, error) {
	return func(ctx context.Context, input common.ActivityInput) (common.ActivityOutput, error) {
		panic("implement me")
	}
}
