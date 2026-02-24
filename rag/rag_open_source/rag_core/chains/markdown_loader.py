import os
import logging
from typing import List, Optional, cast
import re
from langchain_core.documents import Document

from utils.file_encoding_utils import detect_encodings
from pathlib import Path
from typing import Optional, Union
from langchain_community.document_loaders.base import BaseLoader

logger = logging.getLogger(__name__)

class MarkdownLoader(BaseLoader):
    def __init__(
        self,
        file_path: Union[str, Path],
        encoding: Optional[str] = None,
        autodetect_encoding: bool = False,
    ):
        """Initialize with file path."""
        self.file_path = file_path
        self.encoding = encoding
        self.autodetect_encoding = autodetect_encoding


    def markdown_to_tups(self, markdown_text: str) -> list[tuple[Optional[str], str]]:
        """Convert a markdown file to a dictionary.

        The keys are the headers and the values are the text under each header.

        """
        markdown_tups: list[tuple[Optional[str], str]] = []
        lines = markdown_text.split("\n")

        current_header = None
        current_text = ""
        code_block_flag = False

        for line in lines:
            if line.startswith("```"):
                code_block_flag = not code_block_flag
                current_text += line + "\n"
                continue
            if code_block_flag:
                current_text += line + "\n"
                continue
            header_match = re.match(r"^#+\s", line)
            if header_match:
                if current_header is not None:
                    markdown_tups.append((current_header, current_text))

                current_header = line
                current_text = ""
            else:
                current_text += line + "\n"
        markdown_tups.append((current_header, current_text))

        if current_header is not None:
            # pass linting, assert keys are defined
            markdown_tups = [
                (re.sub(r"#", "", cast(str, key)).strip(), re.sub(r"<.*?>", "", value)) for key, value in markdown_tups
            ]
        else:
            markdown_tups = [(key, re.sub("\n", "", value)) for key, value in markdown_tups]

        return markdown_tups

    def markdown_parse_tups(self) -> list[tuple[Optional[str], str]]:
        """Parse file into tuples."""
        text = ""
        try:
            with open(self.file_path, encoding=self.encoding) as f:
                text = f.read()
        except UnicodeDecodeError as e:
            if self.autodetect_encoding:
                detected_encodings = detect_encodings(self.file_path)
                for encoding in detected_encodings:
                    # logger.debug(f"Trying encoding: {encoding.encoding}")
                    logger.info(f"Trying encoding: {encoding.encoding}")
                    try:
                        with open(self.file_path, encoding=encoding.encoding) as f:
                            text = f.read()
                        break
                    except UnicodeDecodeError:
                        continue
            else:
                raise RuntimeError(f"Error loading {self.file_path}") from e
        except Exception as e:
            raise RuntimeError(f"Error loading {self.file_path}") from e

        return self.markdown_to_tups(text)

    def load(self) -> List[Document]:
        documents = []
        try:
            metadata = {"source": self.file_path}
            markdown_tups = self.markdown_parse_tups()
            for header, value in markdown_tups:
                value = value.strip()
                if header is None:
                    documents.append(Document(page_content=value, metadata=metadata))
                else:
                    documents.append(Document(page_content=f"\n\n{header}\n{value}", metadata=metadata))

        except Exception as e:
            import traceback
            logger.error("====> markdown_parser_logs error %s" % e)
            logger.error(traceback.format_exc())
            raise RuntimeError(f"Error loading {self.file_path}") from e

        return documents


if __name__ == "__main__":

    filepath = "./your_file.md"
    loader = MarkdownLoader(filepath)
    docs = loader.load()
    for doc in docs:
        print(doc)