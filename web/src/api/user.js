import service from '@/utils/request';

const hasLang = true;
import { USER_API } from '@/utils/requestConstants';

// 登录
export const login = data => {
  return service({
    url: `${USER_API}/base/login`,
    method: 'post',
    data,
    hasLang,
  });
};

// 统一认证票据换取本地登录态
export const loginSSOExchange = params => {
  return service({
    url: `${USER_API}/base/sso/exchange`,
    method: 'get',
    params,
    hasLang,
  });
};

// 2FA登录
// 第一级验证：密码
export const login2FA1 = data => {
  return service({
    url: `${USER_API}/base/login/email`,
    method: 'post',
    data,
  });
};
// 第二级验证：验证码
// 邮箱验证码
export const login2FA2Code = data => {
  return service({
    url: `${USER_API}/user/login/email/code`,
    method: 'post',
    data,
  });
};
// 首次登录
export const login2FA2new = data => {
  return service({
    url: `${USER_API}/user/login`,
    method: 'put',
    data,
  });
};
// 非首次登录
export const login2FA2exist = data => {
  return service({
    url: `${USER_API}/user/login`,
    method: 'post',
    data,
  });
};

// 获取图形验证码
export const getImgVerCode = () => {
  return service({
    url: `${USER_API}/base/captcha`,
    method: 'get',
    hasLang,
  });
};

// 邮箱注册验证码发送
export const registerCode = data => {
  return service({
    url: `${USER_API}/base/register/email/code`,
    method: 'post',
    data,
  });
};

// 用户邮箱注册
export const register = data => {
  return service({
    url: `${USER_API}/base/register/email`,
    method: 'post',
    data,
  });
};

// 用户名密码注册
export const registerByUsername = data => {
  return service({
    url: `${USER_API}/base/register`,
    method: 'post',
    data,
  });
};

// 重置密码邮箱验证码发送
export const resetCode = data => {
  return service({
    url: `${USER_API}/base/password/email/code`,
    method: 'post',
    data,
  });
};

// 重置密码
export const reset = data => {
  return service({
    url: `${USER_API}/base/password/email`,
    method: 'post',
    data,
  });
};

export const getLangList = () => {
  return service({
    url: `${USER_API}/base/language/select`,
    method: 'get',
  });
};

export const changeLang = data => {
  return service({
    url: `${USER_API}/user/language`,
    method: 'put',
    data,
  });
};

export const getUserDetail = data => {
  return service({
    url: `${USER_API}/user/info`,
    method: 'get',
    params: data,
  });
};

export const getPermission = data => {
  return service({
    url: `${USER_API}/user/permission`,
    method: 'get',
    params: data,
  });
};

export const restUserPassword = data => {
  return service({
    url: `${USER_API}/user/admin/password`,
    method: 'put',
    data,
  });
};
export const restOwnPassword = data => {
  return service({
    url: `${USER_API}/user/password`,
    method: 'put',
    data,
  });
};

export const restAvatar = (data, config) => {
  return service({
    url: `${USER_API}/user/avatar`,
    method: 'put',
    data,
    config,
  });
};

export const docDownload = () => {
  return service({
    url: `${USER_API}/doc_center`,
    method: 'get',
  });
};

// 公用上传 avatar
export const uploadAvatar = (data, config) => {
  return service({
    url: `${USER_API}/avatar`,
    method: 'post',
    data,
    config,
  });
};

// 平台信息
export const getCommonInfo = () => {
  return service({
    url: `${USER_API}/base/custom`,
    method: 'get',
    hasLang,
  });
};
