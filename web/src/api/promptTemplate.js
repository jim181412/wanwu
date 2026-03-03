import service from '@/utils/request';
import { USER_API } from '@/utils/requestConstants';
//获取自定义prompt详情
export const getPromptTemplateDetail = data => {
  return service({
    url: `${USER_API}/prompt/custom`,
    method: 'get',
    params: data,
  });
};
//获取自定义prompt列表
export const getPromptTemplateList = data => {
  return service({
    url: `${USER_API}/prompt/custom/list`,
    method: 'get',
    params: data,
  });
};

//获取内置prompt列表
export const getPromptBuiltInList = data => {
  return service({
    url: `${USER_API}/prompt/template/list`,
    method: 'get',
    params: data,
  });
};
//获取内置prompt详情
export const getPromptBuiltInDetail = data => {
  return service({
    url: `${USER_API}/prompt/template/detail`,
    method: 'get',
    params: data,
  });
};
