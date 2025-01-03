package common

const (
	GetHostTaskQueue string = "GetHostTaskQueue"
	Download         string = "Download"
)

type FileInfo struct {
	Name string
	Path string
	Data []byte
}

type ReadFileInput struct {
	FileInfo
}

type ReadFileOutput struct {
	FileInfo
}

type ActivityInput struct {
	Arguments map[string]string
}

type ActivityOutput struct {
}
