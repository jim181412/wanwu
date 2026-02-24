import http
from email import generator

from flask import jsonify, request

from callback.services import ali_multi_modal as ali_service
from configs.config import config
from utils.response import BizError

from . import callback_bp

generator = ali_service.AliGenAI()


@callback_bp.route("/wan-t2i/wan2.6-t2i", methods=["POST"])
def wan26_t2i():
    """
    通义万相文生图:wan2.6-t2i
    ---
    tags:
      - Tongyi Multi-Modal
    parameters:
      - in: header
        name: Authorization
        schema:
          type: string
        required: true
        description: 认证 Token (格式 Bearer <token>)
    requestBody:
      required: true
      content:
        application/json:
          schema:
            type: object
            required:
              - prompt
            properties:
              prompt:
                type: string
                description: 正向提示词,用于描述期望生成的图像内容、风格和构图。支持中英文,长度不超过2100个字符.
                example: "一只在太空中飞翔的猫，赛博朋克风格"
              negative_prompt:
                type: string
                description: 反向提示词,用于描述不希望在图像中出现的内容,对画面进行限制。支持中英文,长度不超过500个字符
                default: ""
                example: "模糊，低质量"
              size:
                type: string
                description: 输出图像的分辨率，格式为宽*高。总像素在 [1280*1280, 1440*1440] 之间,推荐的分辨率1280*1280、1104*1472、1472*1104、960*1696、1696*960
                default: "1280*1280"
                example: "1280*1280"
              n:
                type: integer
                description: 生成数量
                default: 1
                example: 1
    responses:
      200:
        description: 生成任务提交成功/生成成功
        content:
          application/json:
            schema:
              type: object
              description: 返回生成结果
    """
    auth_header = request.headers.get("Authorization")
    api_key = get_api_key(auth_header)

    data = request.get_json()
    prompt = data.get("prompt")
    if not prompt:
        raise BizError("missing prompt", code=http.HTTPStatus.BAD_REQUEST)
    negative_prompt = data.get("negative_prompt", "")
    size = data.get("size", "1280*1280")
    n = data.get("n", 1)

    res = generator.image_generate(
        api_key=api_key,
        prompt=prompt,
        model="wan2.6-t2i",
        negative_prompt=negative_prompt,
        size=size,
        n=n,
    )
    return jsonify(res)


@callback_bp.route("/qwen-t2i/qwen-image-max", methods=["POST"])
def qwen_image_max():
    """
    通义千问文生图:qwen-image-max
    ---
    tags:
      - Tongyi Multi-Modal
    parameters:
      - in: header
        name: Authorization
        schema:
          type: string
        required: true
        description: 认证 Token (格式 Bearer <token>)
    requestBody:
      required: true
      content:
        application/json:
          schema:
            type: object
            required:
              - prompt
            properties:
              prompt:
                type: string
                description: 正向提示词,用于描述期望生成的图像内容、风格和构图。支持中英文,长度不超过800个字符.
                example: "一只在太空中飞翔的猫，赛博朋克风格"
              negative_prompt:
                type: string
                description: 反向提示词,用于描述不希望在图像中出现的内容,对画面进行限制。支持中英文,长度不超过500个字符
                default: ""
                example: "模糊，低质量"
              size:
                type: string
                description: 输出图像的分辨率，格式为宽*高。默认分辨率为1664*928,可选分辨率及对应比例为1664*928(默认,16:9)、1472*1104(4:3)、1328*1328(1:1)、1104*1472(3:4)和 928*1664(9:16)。
                default: "1664*928"
                example: "1664*928"
              n:
                type: integer
                description: 生成数量
                default: 1
                example: 1
    responses:
      200:
        description: 生成任务提交成功/生成成功
        content:
          application/json:
            schema:
              type: object
              description: 返回生成结果
    """
    auth_header = request.headers.get("Authorization")
    api_key = get_api_key(auth_header)

    data = request.get_json()
    prompt = data.get("prompt")
    if not prompt:
        raise BizError("missing prompt", code=http.HTTPStatus.BAD_REQUEST)
    negative_prompt = data.get("negative_prompt", "")
    size = data.get("size", "1664*928")
    n = data.get("n", 1)

    res = generator.multi_modal_conversation(
        api_key=api_key,
        prompt=prompt,
        model="qwen-image-max",
        negative_prompt=negative_prompt,
        size=size,
        n=n,
    )
    return jsonify(res)


@callback_bp.route("/qwen-i2i/qwen-image-edit-max", methods=["POST"])
def qwen_image_edit_max():
    """
    通义千问图片编辑: qwen-image-edit-max
    ---
    tags:
      - Tongyi Multi-Modal
    summary: 调用通义千问进行图片编辑 (Image-to-Image)
    parameters:
      - in: header
        name: Authorization
        schema:
          type: string
        required: true
        description: 认证 Token (格式 Bearer <token>)
    requestBody:
      required: true
      content:
        application/json:
          schema:
            type: object
            required:
              - prompt
              - images
            properties:
              prompt:
                type: string
                description: 编辑指令 (正向提示词), 用于描述期望对原图进行的修改内容。
                example: "生成一张符合深度图的图像，遵循以下描述：一辆红色的破旧的自行车停在一条泥泞的小路上，背景是茂密的原始森林"
              images:
                type: array
                items:
                  type: string
                description: 输入图像的 URL 或 Base64 编码数据。支持传入1-3张图像。多图输入时,按照数组顺序定义图像顺序
                example: ["https://help-static-aliyun-doc.aliyuncs.com/file-manage-files/zh-CN/20250925/fpakfo/image36.webp"]
              negative_prompt:
                type: string
                description: 反向提示词, 用于描述不希望在图像中出现的内容。
                default: ""
                example: "模糊，低质量，变形"
              size:
                type: string
                description: 输出图像分辨率格式为“宽*高”（如“1024*1536”，宽高范围[512,2048]），常见比例推荐为：1:1（1024*1024、1536*1536）、2:3（768*1152、1024*1536）、3:2（1152*768、1536*1024）、3:4（960*1280、1080*1440）、4:3（1280*960、1440*1080）、9:16（720*1280、1080*1920）、16:9（1280*720、1920*1080）以及 21:9（1344*576、2048*872）。
                default: "1024*1024"
                example: "1024*1024"
              n:
                type: integer
                description: 生成数量
                default: 1
                example: 1
    responses:
      200:
        description: 生成任务提交成功/生成成功
        content:
          application/json:
            schema:
              type: object
              description: 返回生成结果 (通常包含 task_id 或生成的图片地址)
    """
    auth_header = request.headers.get("Authorization")
    api_key = get_api_key(auth_header)

    data = request.get_json()
    prompt = data.get("prompt")
    if not prompt:
        raise BizError("missing prompt", code=http.HTTPStatus.BAD_REQUEST)
    negative_prompt = data.get("negative_prompt", "")
    size = data.get("size")
    n = data.get("n", 1)
    images = data.get("images")
    if not images or not isinstance(images, list):
        raise BizError("missing images", code=http.HTTPStatus.BAD_REQUEST)

    res = generator.image_generate(
        api_key=api_key,
        prompt=prompt,
        model="qwen-image-edit-max",
        images=images,
        negative_prompt=negative_prompt,
        size=size,
        n=n,
    )
    return jsonify(res)


@callback_bp.route("/wan-i2i/wan2.6-image", methods=["POST"])
def wan26_image():
    """
    通义万相图片编辑: wan2.6-image
    ---
    tags:
      - Tongyi Multi-Modal
    summary: 调用通义进行图片编辑 (Image-to-Image)
    parameters:
      - in: header
        name: Authorization
        schema:
          type: string
        required: true
        description: 认证 Token (格式 Bearer <token>)
    requestBody:
      required: true
      content:
        application/json:
          schema:
            type: object
            required:
              - prompt
              - images
            properties:
              prompt:
                type: string
                description: 编辑指令 (正向提示词), 用于描述期望对原图进行的修改内容。
                example: "参考图1的风格和图2的背景，生成番茄炒蛋"
              images:
                type: array
                items:
                  type: string
                description: 输入图像的 URL 或 Base64 编码数据。支持传入1-3张图像。多图输入时,按照数组顺序定义图像顺序
                example: ["https://cdn.wanx.aliyuncs.com/tmp/pressure/umbrella1.png", "https://img.alicdn.com/imgextra/i3/O1CN01SfG4J41UYn9WNt4X1_!!6000000002530-49-tps-1696-960.webp"]
              negative_prompt:
                type: string
                description: 反向提示词, 用于描述不希望在图像中出现的内容。
                default: ""
                example: "模糊，低质量，变形"
              size:
                type: string
                description: 输出图像分辨率格式为“宽*高”（如“1024*1536”，宽高范围[512,2048]），常见比例推荐为：1:1（1024*1024、1536*1536）、2:3（768*1152、1024*1536）、3:2（1152*768、1536*1024）、3:4（960*1280、1080*1440）、4:3（1280*960、1440*1080）、9:16（720*1280、1080*1920）、16:9（1280*720、1920*1080）以及 21:9（1344*576、2048*872）。
                default: "1280*1280"
                example: "1280*1280"
              n:
                type: integer
                description: 生成数量
                default: 1
                example: 1
    responses:
      200:
        description: 生成任务提交成功/生成成功
        content:
          application/json:
            schema:
              type: object
              description: 返回生成结果 (通常包含 task_id 或生成的图片地址)
    """
    auth_header = request.headers.get("Authorization")
    api_key = get_api_key(auth_header)

    data = request.get_json()
    prompt = data.get("prompt")
    if not prompt:
        raise BizError("missing prompt", code=http.HTTPStatus.BAD_REQUEST)
    negative_prompt = data.get("negative_prompt", "")
    size = data.get("size")
    n = data.get("n", 1)
    images = data.get("images")
    if not images or not isinstance(images, list):
        raise BizError("missing images", code=http.HTTPStatus.BAD_REQUEST)

    res = generator.image_generate(
        api_key=api_key,
        prompt=prompt,
        model="wan2.6-image",
        images=images,
        negative_prompt=negative_prompt,
        size=size,
        n=n,
        enable_interleave=False,  # 图像编辑模式
    )
    return jsonify(res)


@callback_bp.route("/wan-i2v/wan2.6-i2v-flash", methods=["POST"])
def wan26_i2v_flash():
    """
    通义万相图生视频: wan2.6-i2v-flash
    ---
    tags:
      - Tongyi Multi-Modal
    summary: 根据首帧图像和文本提示词，生成一段流畅的视频。
    parameters:
      - in: header
        name: Authorization
        schema:
          type: string
        required: true
        description: 认证 Token (格式 Bearer <token>)
    requestBody:
      required: true
      content:
        application/json:
          schema:
            type: object
            required:
              - image
            properties:
              prompt:
                type: string
                description: 文本提示词。用来描述生成图像中期望包含的元素和视觉特点。
                example: "一幅都市奇幻艺术的场景。一个充满动感的涂鸦艺术角色。一个由喷漆所画成的少年，正从一面混凝土墙上活过来。他一边用极快的语速演唱一首英文rap，一边摆着一个经典的、充满活力的说唱歌手姿势。场景设定在夜晚一个充满都市感的铁路桥下。灯光来自一盏孤零零的街灯，营造出电影般的氛围，充满高能量和惊人的细节。视频的音频部分完全由他的rap构成，没有其他对话或杂音。"
              image:
                type: string
                description: 首帧图像的URL或 Base64 编码数据。图像格式：JPEG、JPG、PNG（不支持透明通道）、BMP、WEBP。
                example: "https://help-static-aliyun-doc.aliyuncs.com/file-manage-files/zh-CN/20250925/wpimhv/rap.png"
              negative_prompt:
                type: string
                description: 负向提示词，描述不希望出现的内容。
                default: ""
                example: "模糊，变形，水印"
              audio_url:
                type: string
                description: (可选) 音频文件的 URL，模型将使用该音频生成视频。格式为wav、mp3。
                example: "https://help-static-aliyun-doc.aliyuncs.com/file-manage-files/zh-CN/20250925/ozwpvi/rap.mp3"
              resolution:
                type: string
                description: 指定生成的视频分辨率档位，用于调整视频的清晰度（总像素），可选值为720P、1080P。默认值为720P
                example: "1080P"
              duration:
                type: integer
                description: 视频时长 (秒)。默认为5秒，范围为2到15秒。
                example: 5
              template:
                type: string
                description: (可选) 视频特效模板的名称。若未填写，表示不使用任何视频特效。
              shot_type:
                type: string
                description: single默认值，输出单镜头视频，multi输出多镜头视频。指定生成视频的镜头类型，即视频是由一个连续镜头还是多个切换镜头组成。当希望严格控制视频的叙事结构（如产品展示用单镜头、故事短片用多镜头），可通过此参数指定。
                example: "single"
    responses:
      200:
        description: 任务提交成功
        content:
          application/json:
            schema:
              type: object
              description: 返回生成任务结果 (Task ID 或 结果 URL)
      400:
        description: 请求参数错误
      401:
        description: 未授权
      500:
        description: 服务器内部错误
    """
    auth_header = request.headers.get("Authorization")
    api_key = get_api_key(auth_header)

    data = request.get_json()
    prompt = data.get("prompt")
    negative_prompt = data.get("negative_prompt", "")

    image = data.get("image")
    if not image:
        raise BizError("missing image", code=http.HTTPStatus.BAD_REQUEST)
    audio_url = data.get("audio_url")
    template = data.get("template")
    resolution = data.get("resolution")
    duration = data.get("duration")
    shot_type = data.get("shot_type")

    res = generator.image_to_video_generate(
        api_key=api_key,
        prompt=prompt,
        model="wan2.6-i2v-flash",
        img_url=image,
        audio_url=audio_url,
        negative_prompt=negative_prompt,
        template=template,
        resolution=resolution,
        duration=duration,
        shot_type=shot_type,
    )
    return jsonify(res)


@callback_bp.route("/wan-i2v/wan2.2-kf2v-flash", methods=["POST"])
def wan22_kf2v_flash():
    """
    通义万相首尾帧生成视频: wan2.2-kf2v-flash
    ---
    tags:
      - Tongyi Multi-Modal
    summary: 根据首帧图像和文本提示词，生成一段流畅的视频。
    parameters:
      - in: header
        name: Authorization
        schema:
          type: string
        required: true
        description: 认证 Token (格式 Bearer <token>)
    requestBody:
      required: true
      content:
        application/json:
          schema:
            type: object
            required:
              - first_frame_image
            properties:
              prompt:
                type: string
                description: 文本提示词。用来描述生成图像中期望包含的元素和视觉特点。
                example: "写实风格，一只黑色小猫好奇地看向天空，镜头从平视逐渐上升，最后俯拍它的好奇的眼神。"
              first_frame_image:
                type: string
                description: 首帧图像的URL或 Base64 编码数据。图像格式：JPEG、JPG、PNG（不支持透明通道）、BMP、WEBP。
                example: "https://wanx.alicdn.com/material/20250318/first_frame.png"
              last_frame_image:
                type: string
                description: 首帧图像的URL或 Base64 编码数据。图像格式：JPEG、JPG、PNG（不支持透明通道）、BMP、WEBP。
                example: "https://wanx.alicdn.com/material/20250318/last_frame.png"
              negative_prompt:
                type: string
                description: 负向提示词，描述不希望出现的内容。
                default: ""
                example: "模糊，变形，水印"
              resolution:
                type: string
                description: 指定生成的视频分辨率档位，用于调整视频的清晰度（总像素），可选值为720P、1080P。默认值为1080P
                example: "1080P"
              duration:
                type: integer
                description: 视频时长 (秒)。固定为5秒，不支持修改。
                example: 5
              template:
                type: string
                description: (可选) 视频特效模板的名称。若未填写，表示不使用任何视频特效。支持flying、dissolve
    responses:
      200:
        description: 任务提交成功
        content:
          application/json:
            schema:
              type: object
              description: 返回生成任务结果 (Task ID 或 结果 URL)
      400:
        description: 请求参数错误 (如缺少 image_url)
      401:
        description: 未授权
      500:
        description: 服务器内部错误
    """
    auth_header = request.headers.get("Authorization")
    api_key = get_api_key(auth_header)

    data = request.get_json()
    prompt = data.get("prompt")
    negative_prompt = data.get("negative_prompt", "")

    first_frame_image = data.get("first_frame_image")
    if not first_frame_image:
        raise BizError("missing first_frame_image ", code=http.HTTPStatus.BAD_REQUEST)
    last_frame_image = data.get("last_frame_image")

    template = data.get("template")
    resolution = data.get("resolution")
    duration = 5  # 该模型固定为5s

    res = generator.first_and_last_image_to_video(
        api_key=api_key,
        prompt=prompt,
        model="wan2.2-kf2v-flash",
        first_frame_url=first_frame_image,
        last_frame_url=last_frame_image,
        negative_prompt=negative_prompt,
        template=template,
        resolution=resolution,
        duration=duration,
    )
    return jsonify(res)


@callback_bp.route("/wan-t2v/wan2.6-t2v", methods=["POST"])
def wan26_t2v():
    """
    通义万相文生视频: wan2.6-t2v
    ---
    tags:
      - Tongyi Multi-Modal
    summary: 通义万相文生视频模型基于文本提示词，生成一段流畅的视频。
    parameters:
      - in: header
        name: Authorization
        schema:
          type: string
        required: true
        description: 认证 Token (格式 Bearer <token>)
    requestBody:
      required: true
      content:
        application/json:
          schema:
            type: object
            required:
              - prompt
            properties:
              prompt:
                type: string
                description: 文本提示词。用来描述生成图像中期望包含的元素和视觉特点。
                example: "一只猫在草地上奔跑，高清，电影质感"
              negative_prompt:
                type: string
                description: 负向提示词，描述不希望出现的内容。
                default: ""
                example: "模糊，变形"
              audio_url:
                type: string
                description: (可选) 音频文件的 URL，模型将使用该音频生成视频。格式为wav、mp3。
                example: "https://example.com/audio.mp3"
              size:
                type: string
                description: 指定生成的视频分辨率，格式为宽*高，默认值为1280*720
                example: "1280*720"
              duration:
                type: integer
                description: 视频时长 (秒)。默认为5秒，范围为2到15秒。
                example: 5
              template:
                type: string
                description: (可选) 视频特效模板的名称。若未填写，表示不使用任何视频特效。
              shot_type:
                type: string
                description: single默认值，输出单镜头视频，multi输出多镜头视频。指定生成视频的镜头类型，即视频是由一个连续镜头还是多个切换镜头组成。当希望严格控制视频的叙事结构（如产品展示用单镜头、故事短片用多镜头），可通过此参数指定。
                example: "single"
    responses:
      200:
        description: 任务提交成功
        content:
          application/json:
            schema:
              type: object
              description: 返回生成任务结果 (Task ID 或 结果 URL)
      400:
        description: 请求参数错误 (如缺少 image_url)
      401:
        description: 未授权
      500:
        description: 服务器内部错误
    """
    auth_header = request.headers.get("Authorization")
    api_key = get_api_key(auth_header)

    data = request.get_json()
    prompt = data.get("prompt")
    if not prompt:
        raise BizError("missing prompt", code=http.HTTPStatus.BAD_REQUEST)

    negative_prompt = data.get("negative_prompt", "")
    audio_url = data.get("audio_url")
    size = data.get("size")
    duration = data.get("duration")
    shot_type = data.get("shot_type")

    res = generator.text_to_video_generate(
        api_key=api_key,
        prompt=prompt,
        model="wan2.6-t2v",
        audio_url=audio_url,
        negative_prompt=negative_prompt,
        size=size,
        duration=duration,
        shot_type=shot_type,
    )
    return jsonify(res)


def get_api_key(auth_header):
    api_key = ""
    # 2. 校验格式是否正确 (Bearer <token>)
    if auth_header and auth_header.startswith("Bearer "):
        # 3. 提取 Token (通过空格分割，取第二部分)
        api_key = auth_header.split(" ")[1]

        return api_key
    else:
        raise BizError("missing api key", code=http.HTTPStatus.UNAUTHORIZED)
