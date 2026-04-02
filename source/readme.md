# readme.md


## 环境准备

### local

go run .

默认使用本地模拟登录，不依赖真实 CAS 账号密码。

如需切换到真实 CAS 模式，请设置：

AUTH_MODE=cas
CAS_SERVER_URL=https://your-cas-server/cas


### 访问地址

http://localhost:9090/go/index


## 域名部署示例

如果你希望把当前应用对外访问地址改成：

http://auth.ncu.dilabs.cn/login

可以直接参考：

`deploy/app.env.auth.ncu.dilabs.cn.example`

`deploy/nginx.auth.ncu.dilabs.cn.conf.example`

推荐配置如下：

LISTEN_ADDR=127.0.0.1:9090
CONTEXT_PATH=/login
APP_SERVER_URL=http://auth.ncu.dilabs.cn/login

说明：

* `LISTEN_ADDR` 是 Go 程序实际监听的本机地址
* `APP_SERVER_URL` 是系统对外生成回调地址、登录地址时使用的“外部可访问地址”
* `CONTEXT_PATH=/login` 后，首页路径会变成 `http://auth.ncu.dilabs.cn/login/`，`/login` 通常会自动跳到 `/login/`

本地先验证域名时，可以先改 hosts：

127.0.0.1 auth.ncu.dilabs.cn

然后加载 Nginx 配置，把域名反向代理到 `127.0.0.1:9090`。

如果要切到真实 CAS，再补：

AUTH_MODE=cas
CAS_SERVER_URL=https://your-cas-server/cas

并在 CAS 平台里把以下地址加入配置：

* 应用访问地址：`http://auth.ncu.dilabs.cn/login/index`
* 单点注销地址：`http://auth.ncu.dilabs.cn/login/sso/slo`
* 匹配地址：`http://auth.ncu.dilabs.cn/login/(.*)`


## 对接文档：

[https://authx-service.dev2.supwisdom.com/docs/authx/index.html#/guide/cas/st/guide](https://authx-service.dev2.supwisdom.com/docs/authx/index.html#/guide/cas/st/guide)



## 示例代码使用说明


### cas

本示例代码的核心为 internal/cas 目录下的 cas.go

#### CAS 登录

cas.Login(casServerHostUrl string, service string, requestParams map[string][]string) (*CasUser, string)

> CAS 登录

1. 判断 request 是否存在 ticket
2. 若不存在，跳转到 CAS 认证的登录页面 /login，并返回 null
3. 若存在，则验证 ticket 是否合法 /serviceValidate
4. 验证成功，返回 CasUser；否则，验证失败，返回 CasUser.EMPTY

> 方法参数

* casServerHostUrl ，CAS认证的服务地址，如： `https://cas.xxx.edu.cn/cas`

* service ， 第三方系统的服务地址，便于 认证服务在完成认证后，重新跳转到第三方系统，传递ticket。

* requestParams ， 请求参数，用于获取 请求参数 ticket

> 方法返回

* *CasUser ， CasUser实例

* string ， CAS认证登录的跳转地址， `https://cas.xxx.edu.cn/cas/login?service=abc`


#### CAS 注销

cas.Logout(casServerHostUrl string, service string, requestParams map[string][]string) (bool, string)


> CAS 注销

1. 判断 request 是否存在 logout
2. 若不存在，跳转到 CAS 认证的注销页面 /logout，并返回 false
3. 若存在，则说明 CAS 认证 已注销成功，返回 true

> 方法参数

* casServerHostUrl ，CAS认证的服务地址，如： `https://cas.xxx.edu.cn/cas`

* service ， 第三方系统的服务地址，便于 认证服务在完成注销后，重新跳转到第三方系统，完成第三方系统的后续注销逻辑。

* requestParams ， 请求参数，用于获取 请求参数 logout

> 方法返回

* bool ， 是否注销成功（isLogout）

* string ， CAS认证注销的跳转地址， `https://cas.xxx.edu.cn/cas/logout?service=abc`


#### 票据校验

cas.ServiceValidate(casServerHostUrl string, service string, ticket string) *CasUser

> 票据校验

1. 获取到 CAS认证 返回的 票据（ticket）后，进行合法性验证
2. 验证成功，返回 CasUser
3. 验证失败，返回 CasUser 空对象

> 方法参数

* casServerHostUrl ，CAS认证的服务地址，如： `https://cas.xxx.edu.cn/cas`

* service ， 第三方系统的服务地址，便于 认证服务在完成注销后，重新跳转到第三方系统，完成第三方系统的后续注销逻辑。**确保和 login 时传入的 service 保持完全一致**

* ticket ， CAS认证返回的票据

> 方法返回

* *CasUser ， CasUser实例


#### 账号在线状态检测

**依赖的认证版本：1.2.11-SNAPSHOT，1.3.7-SNAPSHOT，1.4.6-SNAPSHOT，1.5.3-SNAPSHOT**

cas.UserOnlineDetect(casServerHostUrl string, casUser *CasUser) bool

cas.UserOnlineDetect2(casServerHostUrl string, service string, ticket string, username string) bool

> CAS 在线状态检测

1. 根据当前系统中登录账号的用户名，CAS签发的票据 ticket，CAS登录时的 service， 检测该账号在CAS认证中 当前的在线状态
2. 在线，返回 true，系统无须特殊处理，用户可继续操作业务
3. 否则，返回 false，此时系统根据实际情况，可将当前登录账号进行注销，并跳转到 CAS认证 的登录地址进行登录

> 方法参数

* casServerHostUrl ，CAS认证的服务地址，如： `https://cas.xxx.edu.cn/cas`

* service ， 第三方系统的服务地址，用于CAS认证服务对票据真实性的检测。**确保和 login 时传入的 service 保持完全一致，可以从 CasUser 中获取**

* ticket ， CAS认证返回的票据。**可以从 CasUser 中获取**

* username ， 当前登录账号。**可以从 CasUser 中获取**

> 方法返回

* bool ， 是否在线（isAlive）





### sso

本示例代码的登录、注销对接相关的实现为 sso 目录下的 sso.go

**可以根据自身系统的实际进行修改，或自行实现**


#### 登录地址 /sso/login

业务系统需要登录时，请求该登录地址，将跳转到 CAS认证 完成认证

> 登录逻辑

1. 当请求系统某个页面时，若未登录，则重定向跳转到 ${app.server.url}/sso/login?returnUrl=/
2. 该方法接收到请求后，调用 CasUtil.login ，判断是否存在请求参数 ticket，若不存在，将重定向跳转到 CAS认证进行登录
3. CAS认证 完成登录后，仍会 重定向返回到 该地址 ${app.server.url}/sso/login?returnUrl=/&ticket=ST-1-abc-xxx
4. 该方法再次接收到请求后，调用 CasUtil.login ，判断是否存在请求参数 ticket，若存在，则进行 票据（ticket）校验
5. 票据校验成功，将返回 CasUser对象； 票据校验失败，将返回 空的CasUser对象（即 casUser.isEmpty() == true ）
6. 将 CasUser对象 放入 Session
7. 重定向返回到 returnUrl

> 请求方式

`GET https://app.xxx.edu.cn/sso/login?returnUrl=/`


#### 注销地址 /sso/logout

业务系统需要注销时，请求该注销地址，将跳转到 CAS认证 进行注销

> 注销逻辑

1. 系统需要注销时，请求，或重定向跳转到 ${app.server.url}/sso/logout?returnUrl=/
2. 该方法接收到请求后，调用 CasUtil.logout ，判断是否存在请求参数 logout，若不存在，将重定向跳转到 CAS认证进行注销
3. CAS认证 完成注销后，仍会 重定向返回到 该地址 ${app.server.url}/sso/logout?returnUrl=/&logout=logout
4. 该方法再次接收到请求后，调用 CasUtil.logout ，判断是否存在请求参数 logout，则返回 true
5. 将 CasUser对象 从 Session 删除
6. 重定向返回到 returnUrl

> 请求方式

`GET https://app.xxx.edu.cn/sso/logout?returnUrl=/`


#### 在线状态检测 /sso/userOnlineDetect

业务系统暴露给前端调用的 在线状态检测接口

> 检测逻辑

1. 前端代码，通过 javascript 的 ajax POST 请求（无须任何请求参数）
2. 系统获取到 Session 中的 CasUser对象； 若 不存在，则直接返回 未在线 isAlive: false
3. 否则，调用 CasUtil.userOnlineDetect ， 通过 CAS认证 的 在线状态检测接口 进行检测
4. 根据 返回结果，返回 在线 isAlive: true 或 未在线 isAlive: false（若 isAlive 为 false, 可以先清空 Session

> 请求方式

`POST https://app.xxx.edu.cn/sso/userOnlineDetect`


#### 单点注销地址 /sso/slo

由于在浏览器中打开过许多业务系统，
当某个业务系统调用 CAS认证 的注销时，需要将 其他业务系统的 登录状态 进行清除
故，所有的业务系统须按照规范提供 单点注销地址，完成 本系统的 登录状态 的清除

> 单点注销逻辑

1. 接收到请求后，将 Session 清空
2. 响应 jsonp 回调



## CAS认证对接配置

登录 云平台 - 认证管理 - 应用对接配置，按以下信息添加配置：

**实际项目中，根据实际情况配置**

应用标识，如： authx-demo-cas-go

应用域，如： authx-demo-cas.dev2.supwisdom.com

应用名称，如： CAS认证对接示例（go）

应用访问地址，如： https://authx-demo-cas.dev2.supwisdom.com/go/index

单点注销地址，如： https://authx-demo-cas.dev2.supwisdom.com/go/sso/slo

是否启用，选择 是

启用单点登录，选择 是

启用ID Token，若不需要，则选择 否


认证协议，选择 CAS

匹配地址，如： https://authx-demo-cas.dev2.supwisdom.com/go/(.*)

启用JWT Service Ticket， 选择 否（不能选错）

是否适配认证V4，默认即可
