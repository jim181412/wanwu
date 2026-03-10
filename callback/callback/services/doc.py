import concurrent.futures
import io
import json
import logging
import os
import posixpath
import textwrap

import requests
from docx import Document
from reportlab.lib.pagesizes import A4
from reportlab.lib.units import mm
from reportlab.pdfbase import pdfmetrics
from reportlab.pdfbase.ttfonts import TTFont
from reportlab.pdfgen import canvas

from callback.services import minio as minio_service
from configs.config import config
from extensions.minio import minio_client
from utils.build_prompt import build_docqa_prompt_from_search_list
from utils.log import logger
from utils.response import BizError


def process_documents(query, file_urls, sentence_size, overlap_size):
    """
    解析文档并生成 Prompt
    """
    if not file_urls:
        raise BizError("No file URLs provided.")

    # 统一处理为列表
    file_urls = [file_urls] if isinstance(file_urls, str) else file_urls
    all_docs = []

    with concurrent.futures.ThreadPoolExecutor(max_workers=5) as executor:
        future_to_url = {
            executor.submit(parse_doc, url, sentence_size, overlap_size): url
            for url in file_urls
        }

        for future in concurrent.futures.as_completed(future_to_url):
            url = future_to_url[future]
            try:
                docs = future.result()
                all_docs.extend(docs)
            except Exception as e:
                # 这里可以记录日志
                logger.error(f"解析文档失败 {url}: {str(e)}")

    if not all_docs:
        raise BizError("No document content parsed.")

    # 构建文档列表
    doc_list = [
        {
            "snippet": doc.get("text"),
            "file_name": doc.get("metadata", {}).get("file_name"),
        }
        for doc in all_docs
    ]

    # 构建提示词
    prompt = build_docqa_prompt_from_search_list(query, doc_list)
    return prompt


def generate_file_to_minio(formatted_markdown, filename, to_format="txt"):

    pdfmetrics.registerFont(TTFont("SimHei", "callback/static/simhei.ttf"))

    with io.BytesIO() as file_buffer:
        # 1. 初始化变量
        full_filename = filename + ".txt"

        # 2. 根据格式生成文件内容
        if to_format == "pdf":
            full_filename = filename + ".pdf"

            c = canvas.Canvas(file_buffer, pagesize=A4)
            width, height = A4
            margin = 20 * mm
            line_height = 18
            max_width = width - 2 * margin

            c.setFont("SimHei", 12)
            y = height - margin

            # 简单的换行估算
            max_chars_per_line = int(
                max_width // 6
            )  # 粗略修正：中文字符宽，除以12可能太宽，视具体字号调整

            wrapped_lines = []
            for paragraph in formatted_markdown.splitlines():
                wrapped_lines.extend(textwrap.wrap(paragraph, width=max_chars_per_line))
                wrapped_lines.append("")

            for line in wrapped_lines:
                if y < margin:
                    c.showPage()
                    c.setFont("SimHei", 12)
                    y = height - margin
                c.drawString(margin, y, line)
                y -= line_height

            c.save()

        elif to_format == "docx":
            full_filename = filename + ".docx"
            doc = Document()
            doc.add_paragraph(formatted_markdown)
            doc.save(file_buffer)

        elif to_format == "txt":
            full_filename = filename + ".txt"
            file_buffer.write(formatted_markdown.encode("utf-8"))

        # 3. 上传逻辑
        object_path = minio_service.upload_file_to_minio(file_buffer, full_filename)

        download_link = posixpath.join(
            config.callback_cfg["URL"]["MINIO_DOWNLOAD"], object_path
        )

        # 4. 返回结果
        return full_filename, object_path, download_link


def parse_doc(file_url, sentence_size, overlap_size):
    """
    解析单个文档

    参数:
    file_url (str): 文件URL
    sentence_size (int): 句子大小
    overlap_size (float): 重叠比例
    user_token (str, optional): 用户token

    返回:
    list: 解析后的文档片段列表
    """

    url = config.callback_cfg["URL"]["RAG_DOC_PARSER"]
    payload = json.dumps(
        {
            "url": file_url,
            "sentence_size": sentence_size,
            "overlap_size": overlap_size,
            "separators": [
                "\n\n",
                "\n",
                " ",
                ",",
                "\u200b",  # 零宽空格
                "\uff0c",  # 全角逗号
                "\u3001",  # 顿号
                "\uff0e",  # 全角句号
                "\u3002",  # 句号
                "."
            ],
        }
    )
    headers = {"Content-Type": "application/json;charset=utf-8"}
    response = requests.post(url, headers=headers, data=payload, verify=False)
    docs = response.json().get("docs", [])
    return docs
