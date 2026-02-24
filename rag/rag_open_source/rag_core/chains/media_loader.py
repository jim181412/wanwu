# import pandas as pd
# import logging
# import PyPDF2
from pptx import Presentation

from typing import Union, List, Optional
# from langchain.docstore.document import Document
# from langchain.document_loaders import TextLoader
from langchain_core.documents import Document
from langchain_community.document_loaders import TextLoader
import sys
import os
import logging

from utils import asr_utils
from pathlib import Path
from utils import keyframe_extract

logger = logging.getLogger(__name__)

class MediaLoader(TextLoader):
    def __init__(self, file_path: Union[str, Path], encoding: Optional[str] = None, autodetect_encoding: bool = False, parser_choices: List[str] = None,
                 asr_model_id: Optional[str] = None,
                 multimodal_model_id: Optional[str] = None):
        """Initialize a PDFLoader with file path and additional chunk_type."""
        super().__init__(file_path, encoding, autodetect_encoding)  # 确保调用父类的__init__
        # 如果没有提供parser_choices，则使用默认值["asr"]
        if parser_choices is None:
            parser_choices = ["asr"]
        self.parser_choices = parser_choices
        self.asr_model_id = asr_model_id
        self.multimodal_model_id = multimodal_model_id

    def load(self) -> List[Document]:
        text = ""
        path_obj = Path(self.file_path)
        file_name = path_obj.stem
        file_extension = path_obj.suffix

        try:
            if "asr" in self.parser_choices:
                logger.info("=====MediaLoader,file_name:%s,执行音视频ASR解析转文本" % file_name)
                asr_text = asr_utils.asr_parser_text(self.file_path, self.asr_model_id)
                if asr_text:
                    text = "此视频中音频内容：" + asr_text
            if file_extension in [".mp4", ".mov", ".avi"]:
                logger.info("=====MediaLoader,file_name:%s,执行视频画面关键帧解析转文本" % file_name)
                keyframe_text = keyframe_extract.exact_text(self.file_path, self.parser_choices, self.multimodal_model_id)
                if keyframe_text:
                    text = text + "\n 此视频画面内容：" + keyframe_text

        except Exception as e:
            raise RuntimeError(f"Error loading {self.file_path}") from e

        metadata = {"source": self.file_path}
        return [Document(page_content=text, metadata=metadata)]


if __name__ == "__main__":

    filepath = "./WeChat_20250417165837.mp4"
    parser_choices = ["asr", "ocr"]
    loader = MediaLoader(file_path=filepath,parser_choices=parser_choices)
    docs = loader.load()
    for doc in docs:
        print(doc)