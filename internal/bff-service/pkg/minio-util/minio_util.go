package minio_util

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"path/filepath"
	"strings"
	"time"

	http_client "github.com/UnicomAI/wanwu/pkg/http-client"
	"github.com/UnicomAI/wanwu/pkg/log"
	minio_client "github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/minio/minio-go/v7"
)

// URLType 表示URL的类型
type URLType string

const (
	URLTypeMinioDirect    URLType = "MINIO_DIRECT"    // 直接MinIO URL
	URLTypeMinioPresigned URLType = "MINIO_PRESIGNED" // MinIO预签名URL
	URLTypeOther          URLType = "OTHER"           // 其他URL
)

func getMinioURLType(rawURL string) URLType {
	if IsMinioPresignedURL(rawURL) {
		return URLTypeMinioPresigned
	}
	return URLTypeMinioDirect
}

// MinioUrlToBase64 minio文件地址转base64，自动清理本地临时文件
// @param minioUrl minio文件的完整访问URL路径
// @return 文件的base64编码字符串、错误信息
func MinioUrlToBase64(ctx context.Context, minioUrl string) (base64Str string, base64StrWithPrefix string, err error) {
	// 下载文件二进制流到内存
	fileData, _, err := DownloadFile(ctx, minioUrl)
	if err != nil {
		return "", "", err
	}
	// 解析文件后缀用于拼接Base64前缀
	_, _, fileName := SplitMinioPath(minioUrl)
	ext := filepath.Ext(fileName)
	base64Str, base64StrWithPrefix, err = util.FileData2Base64(fileData, "data:"+strings.TrimPrefix(ext, ".")+";base64")
	if err != nil {
		return "", "", err
	}
	return base64Str, base64StrWithPrefix, nil
}

// DownloadFile 从minio下载文件内容到内存，返回文件二进制字节流
// @param ctx 上下文
// @param minioUrl minio文件的完整访问URL路径
// @return []byte 文件的二进制字节流
// @return string 文件名
// @return error 错误信息
func DownloadFile(ctx context.Context, minioUrl string) ([]byte, string, error) {
	urlType := getMinioURLType(minioUrl)
	if urlType == URLTypeMinioPresigned {
		data, err := DownloadFileDirect(ctx, minioUrl)
		filename := GetFilenameFromMinioURL(minioUrl)
		return data, filename, err
	}
	sourceBucketName, objectName, _ := SplitMinioPath(minioUrl)
	object, err := minio_client.Custom().Cli().GetObject(ctx, sourceBucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		log.Errorf("DownloadFileToLocal get minio object error: %s", err)
		return nil, "", err
	}
	defer func() {
		err2 := object.Close()
		if err2 != nil {
			log.Errorf("DownloadFile close minio object error: %s, path: %s", err2, minioUrl)
		}
	}()

	buf := bytes.NewBuffer(nil)
	_, err = io.Copy(buf, object)

	if err != nil && !util.FileEOF(err) {
		log.Errorf("DownloadFileToLocal copy to buffer error: %s, path: %s", err, minioUrl)
		return nil, "", fmt.Errorf("minio file copy to buffer error: %w", err)
	}
	filename := filepath.Base(objectName)
	return buf.Bytes(), filename, nil
}

func GetFilenameFromMinioURL(minioUrl string) string {
	u, err := url.Parse(minioUrl)
	if err != nil {
		return ""
	}
	return filepath.Base(u.Path)
}

// IsMinioPresignedURL 判断是否是MinIO预签名URL
func IsMinioPresignedURL(rawURL string) bool {
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return false
	}

	// 预签名URL通常包含这些查询参数
	query := parsedURL.Query()
	presignedParams := []string{
		"X-Amz-Algorithm",
		"X-Amz-Credential",
		"X-Amz-Date",
		"X-Amz-Expires",
		"X-Amz-SignedHeaders",
		"X-Amz-Signature",
		"AWSAccessKeyId", // 旧版本参数
		"Signature",      // 旧版本参数
	}

	for _, param := range presignedParams {
		if query.Get(param) != "" {
			return true
		}
	}

	return false
}

// DownloadFileDirect 直连下载minio预签名URL文件到内存，返回文件二进制字节流
// @param ctx 上下文
// @param minioFilePath minio的预签名下载URL
// @return []byte 文件的二进制字节流
// @return error 错误信息
func DownloadFileDirect(ctx context.Context, minioUrl string) ([]byte, error) {
	resp, err := http_client.Default().GetOriResp(ctx, &http_client.HttpRequestParams{
		Url:        minioUrl,
		Timeout:    2 * time.Minute,
		MonitorKey: "http_download_service",
		LogLevel:   http_client.LogParams,
	})
	if err != nil {
		return nil, fmt.Errorf("DownloadFileDirect (%v) err: %v", minioUrl, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("DownloadFileDirect (%v) status: %v", minioUrl, resp.StatusCode)
	}
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("DownloadFileDirect (%v) read body err: %v", minioUrl, err)
	}
	return data, nil
}

// SplitMinioPath 解析minio文件完整URL路径，拆分出对应的存储桶名、对象完整路径、文件名
// @param minioUrl minio文件的完整访问URL路径 (例: https://minio.xxx.com/bucketName/folder1/folder2/test.png)
// @return bucketName 	minio对应的存储桶名称
// @return objectName 	minio存储桶内的文件完整对象路径(含文件夹层级)
// @return fileName 	文件的完整名称(包含后缀)
// @note 入参为空字符串时，返回三个空值；URL解析失败/路径格式不合法时，也返回三个空值
func SplitMinioPath(minioUrl string) (bucketName string, objectName string, fileName string) {
	if len(minioUrl) == 0 {
		return "", "", ""
	}
	u, err := url.Parse(minioUrl)
	if err != nil {
		return "", "", ""
	}
	//此处拿到的path是以"/"开头的，因此split的时候split[0]="",数据从split[1]开始
	split := strings.Split(u.Path, "/")
	totalLen := len(split)
	if totalLen > 2 {
		var buffer bytes.Buffer
		for i := 2; i < totalLen; i++ {
			buffer.WriteString(split[i])
			if i < totalLen-1 {
				buffer.WriteString("/")
			}
		}
		return split[1], buffer.String(), split[totalLen-1]
	}
	return "", "", minioUrl
}
