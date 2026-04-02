package response

import "github.com/UnicomAI/wanwu/internal/bff-service/model/request"

type Login struct {
	UID              string            `json:"uid"`
	Username         string            `json:"username"`
	Token            string            `json:"token"`
	ExpiresAt        int64             `json:"expiresAt"`
	ExpireIn         string            `json:"expireIn"`
	Nickname         string            `json:"nickname"`
	OrgPermission    UserOrgPermission `json:"orgPermission"`    // 用户所在组织权限
	Orgs             []IDName          `json:"orgs"`             // 用户所在组织列表
	Language         Language          `json:"language"`         // 语言
	IsUpdatePassword bool              `json:"isUpdatePassword"` // 是否已更新密码
}

type LoginByEmail struct {
	IsEmailCheck     bool   `json:"isEmailCheck"`
	Token            string `json:"token"`
	IsUpdatePassword bool   `json:"isUpdatePassword"` // 是否已更新密码
}

type Captcha struct {
	Key string `json:"key"` // 客户端key
	B64 string `json:"b64"` // 验证码png图片base64字符串
}

type LogoCustomInfo struct {
	Login         CustomLogin         `json:"login"`         // 登录页标题信息
	Home          CustomHome          `json:"home"`          // 首页标题信息
	Tab           CustomTab           `json:"tab"`           // 标签页信息
	About         CustomAbout         `json:"about"`         // 关于信息
	LinkList      map[string]string   `json:"linkList"`      // 跳转链接列表,key为链接名称,value为URL
	Register      CustomRegister      `json:"register"`      // 注册信息
	ResetPassword CustomResetPassword `json:"resetPassword"` // 重置密码信息
	LoginEmail    CustomLoginEmail    `json:"loginEmail"`    // 邮箱登录信息
	DefaultIcon   CustomDefaultIcon   `json:"defaultIcon"`   // 应用默认图片
}

type CustomLogin struct {
	Background       request.Avatar `json:"background"`       // 登录页背景图
	Logo             request.Avatar `json:"logo"`             // 登录页图标
	LoginButtonColor string         `json:"loginButtonColor"` // 登录按钮颜色
	WelcomeText      string         `json:"welcomeText"`      // 登录页欢迎标词
	PlatformDesc     string         `json:"platformDesc"`     // 平台描述词
	UnifiedAuth      UnifiedAuth    `json:"unifiedAuth"`      // 统一认证配置
}

type UnifiedAuth struct {
	Enabled bool `json:"enabled"`
}

type CustomHome struct {
	Logo            request.Avatar `json:"logo"`            // 首页logo
	Title           string         `json:"title"`           // 平台名称
	BackgroundColor string         `json:"backgroundColor"` // 平台背景色
}

type CustomTab struct {
	Logo  request.Avatar `json:"logo"`  // 标签页图标
	Title string         `json:"title"` // 标签页标题
}

type CustomAbout struct {
	LogoPath  string `json:"logoPath"` // 关于图标路径
	Version   string `json:"version"`
	Copyright string `json:"copyright"` // 版权
}

type CustomRegister struct {
	Email CustomEmail `json:"email"` // 注册邮箱
}

type CustomResetPassword struct {
	Email CustomEmail `json:"email"` // 邮箱
}

type CustomLoginEmail struct {
	Email CustomEmail `json:"email"` // 登录邮箱
}

type CustomDefaultIcon struct {
	RagIcon      string `json:"ragIcon"`
	AgentIcon    string `json:"agentIcon"`
	WorkflowIcon string `json:"workflowIcon"`
	PromptIcon   string `json:"promptIcon"`
	ChatflowIcon string `json:"chatflowIcon"`
}

type CustomEmail struct {
	Status bool `json:"status"`
}

type LanguageSelect struct {
	Languages       []Language `json:"languages"`
	DefaultLanguage Language   `json:"defaultLanguage"`
}

type Language struct {
	Code string `json:"code"` // 语言代码
	Name string `json:"name"` // 语言名称
}
