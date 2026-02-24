import os
import base64
import json
import re
import time
import logging
from typing import List, Union, Any, Dict
from openai import OpenAI
from model_manager.model_config import get_model_configure, LlmModelConfig

logger = logging.getLogger(__name__)

def img2base64(img_path):
    img_format = img_path.split('.')[-1]
    with open(img_path, "rb") as f:
        encoded_image = base64.b64encode(f.read())
        encoded_image_text = encoded_image.decode("utf-8")
        img_base64_str = f"data:image/{img_format};base64,{encoded_image_text}"
        return img_base64_str

def parse_error_to_dict(error) -> Dict[str, Any]:
    """将错误信息转换为字典类型"""
    try:
        # 从错误信息中提取 JSON 部分
        error_str = str(error)
        # 使用正则表达式匹配 '-' 后面的 JSON 字符串
        match = re.search(r'-\s*(\{.*\})', error_str)
        if match:
            json_str = match.group(1)
            return json.loads(json_str)
        # 如果没有匹配到 JSON 格式，返回基本错误信息
        return {
            "error": {
                "message": str(error),
                "type": type(error).__name__,
                "code": getattr(error, 'code', 'unknown')
            }
        }
    except Exception as e:
        # 确保总是返回一个有效的错误字典
        return {
            "error": {
                "message": str(error),
                "parse_error": str(e),
                "type": "error_parse_failed"
            }
        }


def req_unicom_VL_plus(image_path: str,
                       multimodal_model_id: str,
                       prompt: str):
    retries = 0
    max_retries = 3
    model_output = ""
    llm_config = get_model_configure(multimodal_model_id)
    logger.info("=========>req_unicom_VL_plus,modelname:%s,provider:%s" % (llm_config.model_name, llm_config.provider))
    if not llm_config.is_vision_support:
        logger.info(" llm is not support vision,multimodal_model_id:%s" % multimodal_model_id)
        return model_output
    while retries < max_retries:
        try:
            client = OpenAI(
                api_key=llm_config.api_key,
                base_url=llm_config.endpoint_url
            )
            messages = [{
                "role": "user",
                "content": [
                    {"type": "text", "text": prompt},
                    {
                        "type": "image_url",
                        "image_url": {
                            "url": img2base64(image_path)
                        }
                    }
                ]
            }]
            extra_params = {}
            if llm_config.provider == "YuanJing":
                extra_params = {  # 模型其他参数，非必传
                    "api_option": "general"  # general:通用；ocr：多模态ocr；math：拍照答题
                }

            completion = client.chat.completions.create(
                model=llm_config.model_name,
                messages=messages,
                stream=False,
                extra_body=extra_params
            )
            model_output = completion.choices[0].message.content

            logger.info("==========>multi_model_output：%s" % repr(model_output))
            return model_output
        except Exception as e:
            error_dict = parse_error_to_dict(e)
            logger.error(f"\n意外错误: {json.dumps(error_dict, ensure_ascii=False)}")
            retries += 1
            time.sleep(1)
    return model_output