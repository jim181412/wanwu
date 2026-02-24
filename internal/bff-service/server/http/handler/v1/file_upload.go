package v1

import (
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	"github.com/gin-gonic/gin"
)

// CheckFile
//
//	@Tags			common.file
//	@Summary		文件校验
//	@Description	校验分片文件
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CheckFileReq	true	"文件校验参数"
//	@Success		200		{object}	response.Response{data=response.CheckFileResp}
//	@Router			/file/check [get]
func CheckFile(ctx *gin.Context) {
	var req request.CheckFileReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.CheckFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// CheckFileList
//
//	@Tags			common.file
//	@Summary		文件列表校验
//	@Description	校验分片文件列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CheckFileListReq	true	"文件列表校验参数"
//	@Success		200		{object}	response.Response{data=response.CheckFileListResp}
//	@Router			/file/check/list [get]
func CheckFileList(ctx *gin.Context) {
	var req request.CheckFileListReq
	if !gin_util.BindQuery(ctx, &req) {
		return
	}
	resp, err := service.CheckFileList(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// UploadFile
//
//	@Tags			common.file
//	@Summary		文件上传
//	@Description	分片文件上传
//	@Security		JWT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			fileName	formData	string	true	"原始文件名"
//	@Param			sequence	formData	int		true	"分片文件序号"
//	@Param			chunkName	formData	string	true	"上传批次标识"
//	@Param			files		formData	file	true	"文件"
//	@Success		200			{object}	response.Response{data=response.UploadFileResp}
//	@Router			/file/upload [post]
func UploadFile(ctx *gin.Context) {
	var req request.UploadFileReq
	if !gin_util.BindForm(ctx, &req) {
		return
	}
	resp, err := service.UploadFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// MergeFile
//
//	@Tags			common.file
//	@Summary		文件合并
//	@Description	合并分片文件，并上传minio
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.MergeFileReq	true	"文件合并参数"
//	@Success		200		{object}	response.Response{data=response.MergeFileResp}
//	@Router			/file/merge [post]
func MergeFile(ctx *gin.Context) {
	var req request.MergeFileReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.MergeFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// CleanFile
//
//	@Tags			common.file
//	@Summary		文件清除
//	@Description	清除已上传的分片文件
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.CleanFileReq	true	"文件清除参数"
//	@Success		200		{object}	response.Response{data=response.CleanFileResp}
//	@Router			/file/clean [post]
func CleanFile(ctx *gin.Context) {
	var req request.CleanFileReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.CleanFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// DeleteFile
//
//	@Tags			common.file
//	@Summary		文件删除
//	@Description	删除已上传的文件
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.DeleteFileReq	true	"文件删除请求参数"
//	@Success		200		{object}	response.Response{data=response.DeleteFileResp}
//	@Router			/file/delete [delete]
func DeleteFile(ctx *gin.Context) {
	var req request.DeleteFileReq
	if !gin_util.Bind(ctx, &req) {
		return
	}
	resp, err := service.DeleteFile(ctx, &req)
	gin_util.Response(ctx, resp, err)
}

// DirectUploadFiles
//
//	@Tags			common
//	@Summary		文件上传
//	@Description	文件上传
//	@Security		JWT
//	@Accept			multipart/form-data
//	@Produce		json
//	@Param			fileName	formData	string	true	"原始文件名"
//	@Param			file		formData	file	true	"文件"
//	@Success		200			{object}	response.Response{data=response.DirectUploadFilesResp}
//	@Router			/file/upload/direct [post]
func DirectUploadFiles(ctx *gin.Context) {
	var req request.DirectUploadFilesReq
	if !gin_util.BindForm(ctx, &req) {
		return
	}
	resp, err := service.DirectUploadFiles(ctx, &req)
	gin_util.Response(ctx, resp, err)
}
