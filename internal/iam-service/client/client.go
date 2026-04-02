package client

import (
	"context"

	errs "github.com/UnicomAI/wanwu/api/proto/err-code"

	"github.com/UnicomAI/wanwu/internal/iam-service/client/model"
	"github.com/UnicomAI/wanwu/internal/iam-service/client/orm"
	"github.com/gromitlee/access/pkg/perm"
)

type IClient interface {

	// --- user ---

	GetAdminUser(ctx context.Context) (uint32, error)

	GetUser(ctx context.Context, userID, orgID uint32) (*orm.UserInfo, *errs.Status)
	GetUsers(ctx context.Context, orgID uint32, name string, offset, limit int32) ([]*orm.UserInfo, int64, *errs.Status)
	SelectUsersNotInOrg(ctx context.Context, orgID uint32, name string) ([]orm.IDName, *errs.Status)
	SelectUsersByUserIDs(ctx context.Context, userIDs []uint32) ([]orm.IDName, *errs.Status)

	CreateUser(ctx context.Context, user *model.User, orgID uint32, roleIDs []uint32) (uint32, *errs.Status)
	UpdateUser(ctx context.Context, user *model.User, orgID uint32, roleIDs []uint32) *errs.Status
	DeleteUser(ctx context.Context, userID uint32) *errs.Status
	UpdateUserAvatar(ctx context.Context, userID uint32, key string) *errs.Status

	ChangeUserStatus(ctx context.Context, userID, orgID uint32, status bool) *errs.Status
	UpdateUserPassword(ctx context.Context, userID uint32, pwd, newPwd string) *errs.Status
	ResetUserPassword(ctx context.Context, userID uint32, pwd string) *errs.Status

	GetUserPermission(ctx context.Context, userID, orgID uint32) (*orm.Permission, *errs.Status)
	ChangeUserLanguage(ctx context.Context, userID uint32, language string) *errs.Status

	// --- org ---

	GetTopOrg(ctx context.Context) (uint32, error)

	GetOrg(ctx context.Context, orgID uint32) (*orm.OrgInfo, *errs.Status)
	GetOrgs(ctx context.Context, parentID uint32, name string, offset, limit int32) ([]*orm.OrgInfo, int64, *errs.Status)
	SelectOrgs(ctx context.Context, userID uint32) ([]orm.IDName, *errs.Status)
	GetOrgByOrgIDs(ctx context.Context, orgIDs []uint32) ([]orm.IDName, *errs.Status)
	GetOrgAndSubOrgSelectByUser(ctx context.Context, userID, orgID uint32) ([]orm.IDName, *errs.Status)
	GetFirstClassOrgAndSubs(ctx context.Context, userID, orgID uint32) ([]orm.IDName, *errs.Status)

	CreateOrg(ctx context.Context, org *model.Org) (uint32, *errs.Status)
	UpdateOrg(ctx context.Context, org *model.Org) *errs.Status
	DeleteOrg(ctx context.Context, orgID uint32) *errs.Status

	ChangeOrgStatus(ctx context.Context, orgID uint32, status bool) *errs.Status
	AddOrgUser(ctx context.Context, orgID, userID, roleID uint32) *errs.Status
	RemoveOrgUser(ctx context.Context, orgID, userID uint32) *errs.Status

	// --- role ---

	GetAdminRole(ctx context.Context) (uint32, error)

	GetRole(ctx context.Context, orgID, roleID uint32) (*orm.RoleInfo, *errs.Status)
	GetRoles(ctx context.Context, orgID uint32, name string, offset, limit int32) ([]*orm.RoleInfo, int64, *errs.Status)
	SelectRoles(ctx context.Context, orgID uint32) ([]orm.RoleIDName, *errs.Status)

	CreateRole(ctx context.Context, orgID, creatorID uint32, name, remark string, perms []perm.Perm) (uint32, error)
	UpdateRole(ctx context.Context, orgID, roleID uint32, name, remark string, perms []perm.Perm) *errs.Status
	DeleteRole(ctx context.Context, orgID, roleID uint32) *errs.Status

	ChangeRoleStatus(ctx context.Context, orgID, roleID uint32, status bool) *errs.Status

	// --- perm ---

	CheckUserOK(ctx context.Context, userID uint32, genTokenAt int64) (bool, string, int64, *errs.Status)
	CheckUserPerm(ctx context.Context, userID uint32, genTokenAt int64, orgID uint32, oneOfPerms []perm.Perm) (bool, bool, string, int64, *errs.Status)

	// --- captcha ---

	RefreshCaptcha(ctx context.Context, key, code string) *errs.Status
	CheckCaptcha(ctx context.Context, key, code string) *errs.Status

	// --- login ---

	Login(ctx context.Context, username, password, language string) (*orm.UserInfo, *orm.Permission, *errs.Status)
	LoginSendEmailCode(ctx context.Context, email string) *errs.Status
	LoginByEmail(ctx context.Context, username, password string) (*orm.EmailLoginInfo, *errs.Status)
	LoginEmailCheck(ctx context.Context, userID uint32, email, code, language string) (*orm.UserInfo, *orm.Permission, *errs.Status)
	ChangeUserPasswordByEmail(ctx context.Context, userID uint32, OldPassword, NewPassword, email, code, language string) (*orm.UserInfo, *orm.Permission, *errs.Status)

	// --- register ---

	RegisterByUsername(ctx context.Context, username, password string) *errs.Status
	RegisterSendEmailCode(ctx context.Context, username, email string) *errs.Status
	RegisterByEmail(ctx context.Context, username, email, code string) *errs.Status

	// --- reset password ---

	ResetPasswordSendEmailCode(ctx context.Context, email string) *errs.Status
	ResetPasswordByEmail(ctx context.Context, email, password, code string) *errs.Status

	// --- oauth app ---

	CreateOauthApp(ctx context.Context, req *model.OauthApp) *errs.Status
	DeleteOauthApp(ctx context.Context, clientID string) *errs.Status
	UpdateOauthApp(ctx context.Context, req *model.OauthApp) *errs.Status
	GetOauthAppList(ctx context.Context, userID uint32, name string, offset, limit int32) ([]*model.OauthApp, int64, *errs.Status)
	UpdateOauthAppStatus(ctx context.Context, clientID string, status bool) *errs.Status
	GetOauthApp(ctx context.Context, clientID string) (*model.OauthApp, *errs.Status)
}
