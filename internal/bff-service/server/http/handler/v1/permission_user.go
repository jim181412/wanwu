package v1

import (
	"fmt"

	err_code "github.com/UnicomAI/wanwu/api/proto/err-code"
	"github.com/UnicomAI/wanwu/internal/bff-service/model/request"
	"github.com/UnicomAI/wanwu/internal/bff-service/service"
	gin_util "github.com/UnicomAI/wanwu/pkg/gin-util"
	grpc_util "github.com/UnicomAI/wanwu/pkg/grpc-util"
	"github.com/gin-gonic/gin"
)

// CreateUser
//
//	@Tags			permission.user
//	@Summary		创建用户
//	@Description	创建用户，同时加入X-Org-Id组织；在系统视角下创建用户，不加入任何组织，也不能分配角色
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UserCreate	true	"用户信息"
//	@Success		200		{object}	response.Response{data=response.UserID}
//	@Router			/user [post]
func CreateUser(ctx *gin.Context) {
	var req request.UserCreate

	fmt.Print("------当前的创建用户request body-----")
	fmt.Println(ctx.Request)
	fmt.Println("---------")
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_add_other"))
		return
	}
	resp, err := service.CreateUser(ctx, getUserID(ctx), getOrgID(ctx), &req)
	gin_util.Response(ctx, resp, err)
}

// ChangeUser
//
//	@Tags			permission.user
//	@Summary		编辑用户
//	@Description	编辑X-Org-Id组织的用户；在系统视角下编辑用户，不能分配角色
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UserUpdate	true	"用户信息"
//	@Success		200		{object}	response.Response
//	@Router			/user [put]
func ChangeUser(ctx *gin.Context) {
	var req request.UserUpdate
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_change_other"))
		return
	}
	err := service.ChangeUser(ctx, getOrgID(ctx), &req)
	gin_util.Response(ctx, nil, err)
}

// DeleteUser
//
//	@Tags			permission.user
//	@Summary		删除用户
//	@Description	从X-Org-Id组织将用户移除；在系统视角下为删除用户
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.UserID	true	"用户ID"
//	@Success		200		{object}	response.Response
//	@Router			/user [delete]
func DeleteUser(ctx *gin.Context) {
	var req request.UserID
	if !gin_util.Bind(ctx, &req) {
		return
	}
	// delete
	if isSystem(ctx) {
		if !isAdmin(ctx) {
			gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_delete"))
			return
		}
		err := service.DeleteUser(ctx, req.UserID)
		gin_util.Response(ctx, nil, err)
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_delete_other"))
		return
	}
	// remove from org
	err := service.RemoveOrgUser(ctx, getOrgID(ctx), req.UserID)
	gin_util.Response(ctx, nil, err)
}

// GetUserList
//
//	@Tags			permission.user
//	@Summary		获取用户列表
//	@Description	获取X-Org-Id组织的用户列表；在系统视角下获取系统内全部用户列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name		query		string	false	"用户名(模糊查询)"
//	@Param			pageNo		query		int		true	"页面编号，从1开始"
//	@Param			pageSize	query		int		true	"单页数量，从1开始"
//	@Success		200			{object}	response.Response{data=response.PageResult{list=[]response.UserInfo}}
//	@Router			/user/list [get]
func GetUserList(ctx *gin.Context) {
	resp, err := service.GetUserList(ctx, getOrgID(ctx), ctx.Query("name"), getPageNo(ctx), getPageSize(ctx))
	gin_util.Response(ctx, resp, err)
}

// ChangeUserStatus
//
//	@Tags		permission.user
//	@Summary	修改用户状态
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.UserStatus	true	"用户信息"
//	@Success	200		{object}	response.Response
//	@Router		/user/status [put]
func ChangeUserStatus(ctx *gin.Context) {
	var req request.UserStatus
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_change_status"))
		return
	}
	err := service.ChangeUserStatus(ctx, req.UserID.UserID, getOrgID(ctx), req.Status)
	gin_util.Response(ctx, nil, err)
}

// ChangeUserPassword
//
//	@Tags		permission.user
//	@Summary	修改用户密码（by 个人）
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.UserPassword	true	"用户信息"
//	@Success	200		{object}	response.Response
//	@Router		/user/password [put]
func ChangeUserPassword(ctx *gin.Context) {
	var req request.UserPassword
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if req.UserID.UserID != getUserID(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_change_other_password"))
		return
	}
	err := service.ChangeUserPassword(ctx, req.UserID.UserID, req.OldPassword, req.NewPassword)
	gin_util.Response(ctx, nil, err)
}

// AdminChangeUserPassword
//
//	@Tags		permission.user
//	@Summary	重置用户密码（by 管理员）
//	@Security	JWT
//	@Accept		json
//	@Produce	json
//	@Param		data	body		request.UserPasswordByAdmin	true	"用户信息"
//	@Success	200		{object}	response.Response
//	@Router		/user/admin/password [put]
func AdminChangeUserPassword(ctx *gin.Context) {
	var req request.UserPasswordByAdmin
	if !gin_util.Bind(ctx, &req) {
		return
	}
	if !isAdmin(ctx) {
		gin_util.Response(ctx, nil, grpc_util.ErrorStatusWithKey(err_code.Code_BFFGeneral, "bff_user_cannot_change_other_password"))
		return
	}
	err := service.AdminChangeUserPassword(ctx, req.UserID.UserID, req.Password)
	gin_util.Response(ctx, nil, err)
}

// GetOrgUserNotSelect
//
//	@Tags			permission.user
//	@Summary		获取不在组织中用户列表（用于下拉选择）
//	@Description	获取非X-Org-Id组织的用户列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			name	query		string	false	"用户名(模糊查询)"
//	@Success		200		{object}	response.Response{data=response.Select}
//	@Router			/org/other/select [get]
func GetOrgUserNotSelect(ctx *gin.Context) {
	resp, err := service.GetOrgUserNotSelect(ctx, getOrgID(ctx), ctx.Query("name"))
	gin_util.Response(ctx, resp, err)
}

// GetRoleSelect
//
//	@Tags			permission.user
//	@Summary		获取组织角色列表（用于下拉选择）
//	@Description	获取X-Org-Id组织的角色列表
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Success		200	{object}	response.Response{data=response.Select}
//	@Router			/role/select [get]
func GetRoleSelect(ctx *gin.Context) {
	resp, err := service.GetRoleSelect(ctx, getOrgID(ctx))
	gin_util.Response(ctx, resp, err)
}

// AddOrgUser
//
//	@Tags			permission.user
//	@Summary		邀请用户加入组织
//	@Description	增加X-Org-Id组织的用户
//	@Security		JWT
//	@Accept			json
//	@Produce		json
//	@Param			data	body		request.OrgUserAdd	true	"用户-角色"
//	@Success		200		{object}	response.Response
//	@Router			/org/user [post]
func AddOrgUser(ctx *gin.Context) {
	var req request.OrgUserAdd
	if !gin_util.Bind(ctx, &req) {
		return
	}
	err := service.AddOrgUser(ctx, getOrgID(ctx), req.UserID.UserID, req.RoleID)
	gin_util.Response(ctx, nil, err)
}
