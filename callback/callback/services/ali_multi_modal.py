# -*- coding: utf-8 -*-
import dashscope
from dashscope import ImageSynthesis, MultiModalConversation, VideoSynthesis
from dashscope.aigc.image_generation import ImageGeneration
from dashscope.api_entities.dashscope_response import Message

from utils.log import logger


class AliGenAI:
    def __init__(self, region="cn"):
        """
        初始化阿里云生图客户端
        :param api_key: 阿里云百炼 API Key
        :param region: 地域 ('cn' 为北京, 'intl' 为新加坡)
        """

        # 设置 API URL
        if region == "intl":
            dashscope.base_http_api_url = "https://dashscope-intl.aliyuncs.com/api/v1"
        else:
            dashscope.base_http_api_url = "https://dashscope.aliyuncs.com/api/v1"

    # 适用范围：图片生成，旧版接口。
    def image_generate_legacy(
        self, api_key, prompt, image_urls=None, model="wan2.5-t2i-preview", **kwargs
    ):
        """
        调用 ImageSynthesis，适用模型：仅适合wan2.5及以下版本模型以及qwen-image-plus、qwen-image
        """

        n = kwargs.get("n", 1)
        size = kwargs.get("size", "1280*1280")
        negative_prompt = kwargs.get("negative_prompt", "")

        try:
            rsp = ImageSynthesis.call(
                api_key=api_key,
                model=model,
                prompt=prompt,
                images=image_urls,
                negative_prompt=negative_prompt,
                n=n,
                size=size,
                prompt_extend=True,
                watermark=True,
            )
            return rsp
        except Exception as e:
            print(f"Wan 模型调用异常: {e}")
            return None

    # 适用范围：图片生成，新版接口。
    def image_generate(
        self, api_key, prompt, images=None, model="qwen-image-plus-2026-01-09", **kwargs
    ):
        """
        调用 ImageGeneration (适用于同步Api模型)
        增加了 image_url 参数以支持图生图/控制生图 (如 qwen-image-edit-max)
        """

        n = kwargs.get("n", 1)
        size = kwargs.get("size", "1280*1280")
        negative_prompt = kwargs.get("negative_prompt", " ")

        # 构造 Content 列表
        content_list = []

        logger.info(f"image_urls: {images}")

        # 1. 如果有输入图片，先添加图片信息,包含1-3张图像
        if images and isinstance(images, list):
            for img in images:  # 最多支持3张图片
                content_list.append({"image": img})

        # 2. 添加文本提示词
        content_list.append({"text": prompt})

        logger.info(f"content_list: {content_list}")

        # 构造 Message 对象
        message = Message(role="user", content=content_list)

        try:
            rsp = ImageGeneration.call(
                model=model,
                api_key=api_key,
                messages=[message],
                negative_prompt=negative_prompt,  # 对应 curl 中的 " "
                prompt_extend=True,  # 是否开启提示词优化
                watermark=True,  # 是否添加水印
                n=n,
                size=size,
            )
            return rsp
        except Exception as e:
            print(f"Qwen 模型调用异常: {e}")
            return None

    def multi_modal_conversation(
        self, api_key, prompt, image_url=None, model="qwen-image-max", **kwargs
    ):
        """
        调用多模态对话接口
        """
        n = kwargs.get("n", 1)
        size = kwargs.get("size", "1280*1280")
        negative_prompt = kwargs.get("negative_prompt", "")

        # 构造 Content 列表
        content_list = []

        # 1. 如果有输入图片，先添加图片信息,包含1-3张图像
        if image_url and isinstance(image_url, list):
            for url in image_url:  # 最多支持3张图片
                content_list.append({"image": url})

        # 2. 添加文本提示词
        content_list.append({"text": prompt})

        # 构造 Message 对象
        messages = [{"role": "user", "content": content_list}]

        try:
            rsp = MultiModalConversation.call(
                api_key=api_key,
                model=model,
                messages=messages,
                negative_prompt=negative_prompt,
                n=n,
                size=size,
                prompt_extend=True,
                watermark=True,
            )
            return rsp
        except Exception as e:
            print(f"Wan 模型调用异常: {e}")
            return None

    def image_to_video_generate(self, **kwargs):
        """
        调用图片生成视频接口
        """
        api_key = kwargs.get("api_key")
        if not api_key:
            raise ValueError("missing api_key")

        prompt = kwargs.get("prompt")
        if not prompt:
            raise ValueError("missing prompt ")

        model = kwargs.get("model", "default-video-model")
        if not model:
            raise ValueError("missing model ")

        img_url = kwargs.get("img_url")
        if not img_url:
            raise ValueError("missing img_url ")
        # 音频链接
        audio_url = kwargs.get("audio_url")
        # 视频的清晰度
        resolution = kwargs.get("resolution", " 720P")
        # 生成视频的时长
        duration = kwargs.get("duration", 5)

        negative_prompt = kwargs.get("negative_prompt", "")
        try:
            resp = VideoSynthesis.call(
                api_key=api_key,
                model=model,
                prompt=prompt,
                img_url=img_url,
                audio_url=audio_url,
                resolution=resolution,
                duration=duration,
                prompt_extend=True,
                watermark=True,
                negative_prompt=negative_prompt,
            )
            return resp
        except Exception as e:
            print(f"视频生成异常: {e}")
            return e

    def first_and_last_image_to_video(self, **kwargs):
        """
        调用图片生成视频接口
        """
        api_key = kwargs.get("api_key")
        if not api_key:
            raise ValueError("missing api_key")

        prompt = kwargs.get("prompt")
        if not prompt:
            raise ValueError("missing prompt ")

        model = kwargs.get("model", "default-video-model")
        if not model:
            raise ValueError("missing model ")

        first_frame_url = kwargs.get("first_frame_url")
        if not first_frame_url:
            raise ValueError("missing first_frame_url ")

        last_frame_url = kwargs.get("last_frame_url")

        # 视频的清晰度
        resolution = kwargs.get("resolution", " 720P")
        # 生成视频的时长
        duration = kwargs.get("duration", 5)

        negative_prompt = kwargs.get("negative_prompt", "")
        try:
            resp = VideoSynthesis.call(
                api_key=api_key,
                model=model,
                prompt=prompt,
                first_frame_url=first_frame_url,
                last_frame_url=last_frame_url,
                resolution=resolution,
                duration=duration,
                prompt_extend=True,
                watermark=True,
                negative_prompt=negative_prompt,
            )
            return resp
        except Exception as e:
            print(f"视频生成异常: {e}")
            return e

    def text_to_video_generate(self, **kwargs):
        """
        调用文本生成视频接口
        """
        api_key = kwargs.get("api_key")
        if not api_key:
            raise ValueError("missing api_key")
        prompt = kwargs.get("prompt")
        if not prompt:
            raise ValueError("missing prompt ")
        model = kwargs.get("model")
        if not model:
            raise ValueError("missing model ")
        negative_prompt = kwargs.get("negative_prompt", "")
        audio_url = kwargs.get("audio_url")
        size = kwargs.get("size", "1280*720")
        duration = kwargs.get("duration", 5)

        rsp = VideoSynthesis.call(
            api_key=api_key,
            model=model,
            prompt=prompt,
            audio_url=audio_url,
            size=size,
            duration=duration,
            negative_prompt=negative_prompt,
            prompt_extend=True,
            watermark=True,
        )
        print(rsp)
        if rsp.status_code == 200:
            print(rsp.output.video_url)
        else:
            print(
                "Failed, status_code: %s, code: %s, message: %s"
                % (rsp.status_code, rsp.code, rsp.message)
            )

        return rsp
