package service

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strings"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"
	mcp_service "github.com/UnicomAI/wanwu/api/proto/mcp-service"
	"github.com/UnicomAI/wanwu/internal/bff-service/config"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/response"
	minio_util "github.com/UnicomAI/wanwu/internal/bff-service/pkg/minio-util"
	"github.com/UnicomAI/wanwu/pkg/constant"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/UnicomAI/wanwu/pkg/minio"
	"github.com/UnicomAI/wanwu/pkg/util"
	"github.com/gin-gonic/gin"
)

const (
	customSkillFileType = ".zip"
)

func CreateCustomSkill(ctx *gin.Context, userId, orgId string, avatarKey, author, zipUrl, saveId, sourceType string) (*response.CustomSkillIDResp, error) {
	var objectPath, markdownContent, skillName, skillDesc string

	if zipUrl != "" {
		// 下载文件
		data, err := minio_util.DownloadFileDirect(ctx.Request.Context(), zipUrl)
		if err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("download skill zip err: %v", err))
		}

		// 解压并查找SKILL.md文件，提取name和description
		mdContent, fm, err := util.ExtractSkillMarkdownFromZip(data)
		if err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFSkillParse, err.Error())
		}
		markdownContent = mdContent

		// 如果从markdown中提取到了name和desc，使用这些值
		skillName = fm.Name
		skillDesc = fm.Description

		fileName, _, err := minio.UploadFileCommon(ctx.Request.Context(), bytes.NewReader(data), customSkillFileType, -1, true)
		if err != nil {
			return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, err.Error())
		}
		// 构建完整的相对路径：file-upload/file-not-expire/xxx.zip
		objectPath = filepath.Join(minio.BucketFileUpload, minio.DirFileNotExpire, fileName)
	}

	createResp, err := mcp.CustomSkillCreate(ctx.Request.Context(), &mcp_service.CustomSkillCreateReq{
		Name:       skillName,
		Avatar:     avatarKey,
		Author:     author,
		Desc:       skillDesc,
		ObjectPath: objectPath,
		Markdown:   markdownContent,
		SaveId:     saveId,
		SourceType: sourceType,
		Identity:   &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}

	return &response.CustomSkillIDResp{SkillId: createResp.SkillId}, nil
}

func GetCustomSkill(ctx *gin.Context, userId, orgId, skillId string) (*response.CustomSkillDetail, error) {
	resp, err := mcp.CustomSkillGet(ctx.Request.Context(), &mcp_service.CustomSkillGetReq{
		SkillId:  skillId,
		Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}

	return &response.CustomSkillDetail{
		SkillDetail: response.SkillDetail{
			SkillId:       resp.SkillId,
			Name:          resp.Name,
			Avatar:        cacheSkillAvatar(ctx, resp.Avatar),
			Author:        resp.Author,
			Desc:          resp.Desc,
			SkillMarkdown: config.FixFrontMatterFormat(resp.Markdown),
		},
		ZipUrl: buildAccessFilePath(resp.ObjectPath),
	}, nil
}

func DeleteCustomSkill(ctx *gin.Context, skillId string) error {
	_, err := mcp.CustomSkillDelete(ctx.Request.Context(), &mcp_service.CustomSkillDeleteReq{
		SkillId: skillId,
	})
	return err
}

func GetCustomSkillList(ctx *gin.Context, userId, orgId, name string) (*response.ListResult, error) {
	resp, err := mcp.CustomSkillGetList(ctx.Request.Context(), &mcp_service.CustomSkillGetListReq{
		Name:     name,
		Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}

	customSkillList := make([]*response.CustomSkillDetail, 0, len(resp.List))
	for _, skill := range resp.List {
		customSkillList = append(customSkillList, toCustomSkill(ctx, skill))
	}

	return &response.ListResult{
		List:  customSkillList,
		Total: resp.Total,
	}, nil
}

func toCustomSkill(ctx *gin.Context, skill *mcp_service.CustomSkill) *response.CustomSkillDetail {
	if skill == nil {
		return nil
	}
	return &response.CustomSkillDetail{
		SkillDetail: response.SkillDetail{
			SkillId: skill.SkillId,
			Name:    skill.Name,
			Avatar:  cacheSkillAvatar(ctx, skill.Avatar),
			Author:  skill.Author,
			Desc:    skill.Desc,
		},
		ZipUrl: buildAccessFilePath(skill.ObjectPath),
	}
}

func CheckCustomSkill(ctx *gin.Context, userId, orgId, zipUrl string) (*response.CustomSkillCheckResp, error) {
	// 下载文件
	data, err := minio_util.DownloadFileDirect(ctx.Request.Context(), zipUrl)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFGeneral, fmt.Sprintf("download skill zip err: %v", err))
	}

	// 解压并查找SKILL.md文件，验证zip包是否有效
	_, fm, err := util.ExtractSkillMarkdownFromZip(data)
	if err != nil {
		return nil, grpc_util.ErrorStatus(errs.Code_BFFSkillParse, err.Error())
	}

	return &response.CustomSkillCheckResp{
		Name: fm.Name,
		Desc: fm.Description,
	}, nil
}

func GetSkillSelect(ctx *gin.Context, userId, orgId, name string) (*response.ListResult, error) {
	var allSkills []*response.SkillInfo

	// 内建 skills
	for _, skillsCfg := range config.Cfg().AgentSkills {
		if name != "" && !strings.Contains(skillsCfg.Name, name) {
			continue
		}
		iconUrl := config.Cfg().DefaultIcon.SkillIcon
		if skillsCfg.Avatar != "" {
			iconUrl = skillsCfg.Avatar
		}
		allSkills = append(allSkills, &response.SkillInfo{
			SkillId:   skillsCfg.SkillId,
			SkillName: skillsCfg.Name,
			SkillType: constant.SkillTypeBuiltIn,
			Desc:      skillsCfg.Desc,
			Author:    skillsCfg.Author,
			Avatar:    request.Avatar{Path: iconUrl},
		})
	}

	// 自定义 skills
	customResp, err := mcp.CustomSkillGetList(ctx.Request.Context(), &mcp_service.CustomSkillGetListReq{
		Name:     name,
		Identity: &mcp_service.Identity{UserId: userId, OrgId: orgId},
	})
	if err != nil {
		return nil, err
	}

	for _, skill := range customResp.List {
		allSkills = append(allSkills, &response.SkillInfo{
			SkillId:   skill.SkillId,
			SkillName: skill.Name,
			SkillType: constant.SkillTypeCustom,
			Desc:      skill.Desc,
			Author:    skill.Author,
			Avatar:    cacheSkillAvatar(ctx, skill.Avatar),
		})
	}

	return &response.ListResult{
		List:  allSkills,
		Total: int64(len(allSkills)),
	}, nil
}
