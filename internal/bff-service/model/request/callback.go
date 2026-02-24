package request

type FileUrlConvertBase64Req struct {
	FileUrl      string `form:"fileUrl" json:"fileUrl" validate:"required"` // 文件URL
	AddPrefix    bool   `form:"addPrefix" json:"addPrefix"`                 // 是否添加 data:xxx;base64, 前缀
	CustomPrefix string `form:"customPrefix" json:"customPrefix"`           // 自定义前缀（如 "data:video/mp4;base64,"）
}

func (f *FileUrlConvertBase64Req) Check() error {
	return nil
}
