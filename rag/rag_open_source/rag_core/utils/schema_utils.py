import os
import json
import pandas as pd
import sys
import logging

logger = logging.getLogger(__name__)

def parse_excel_to_schema_json(file_path):
    """
    解析 Excel 文件中的 '类目表' 和 '类目属性表'，输出指定 JSON 结构
    """
    schema = {}
    try:
        # 使用 pd.read_excel 自动推断引擎（支持 .xls 和 .xlsx）
        df_category = pd.read_excel(file_path, sheet_name='类目表')
        df_attribute = pd.read_excel(file_path, sheet_name='类目属性表')
        # 清理列名：去除空格和换行
        df_category.columns = df_category.columns.str.strip()
        df_attribute.columns = df_attribute.columns.str.strip()

        # === 解析 类目表 ===
        category_list = []
        for _, row in df_category.iterrows():
            item = {
                "类名": str(row["类名"]).strip() if pd.notna(row["类名"]) else "",
                "类描述": str(row["类描述"]).strip() if pd.notna(row["类描述"]) else ""
            }
            category_list.append(item)

        # === 解析 类目属性表 ===
        attribute_list = []

        for _, row in df_attribute.iterrows():
            class_name = str(row["类名"]).strip() if pd.notna(row["类名"]) else ""
            attr_name = str(row["属性/关系名"]).strip() if pd.notna(row["属性/关系名"]) else ""

            # 修复说明字段
            key = (class_name, attr_name)

            desc = str(row["属性/关系说明"]).strip() if pd.notna(row["属性/关系说明"]) else ""

            # 处理别名字段（支持多个别名用 | 分隔）
            alias = row["别名(多别名以|隔开)"]
            if pd.isna(alias) or str(alias).strip() == "" or str(alias).lower() == "nan":
                alias_str = ""
            else:
                alias_str = str(alias).strip()

            value_type = str(row["值类型"]).strip() if pd.notna(row["值类型"]) else ""

            attribute_list.append({
                "类名": class_name,
                "属性/关系名": attr_name,
                "属性/关系说明": desc,
                "属性别名(多别名以|隔开)": alias_str,
                "值类型": value_type
            })

        # 构建最终 JSON 结构
        schema = {
            "schema定义": {
                "类目表": category_list,
                "类目属性表": attribute_list
            }
        }
        logger.info("schema:%s" % json.dumps(schema, ensure_ascii=False))
    except Exception as e:
        import traceback
        logger.error(traceback.format_exc())
        logger.error(f"无法读取Excel文件或工作表不存在: {e}")
    return schema


if __name__ == "__main__":

    excel_file = "graph_schema_文物.xlsx"  # 或 "图谱schema定义模板.xls"

    try:
        schema_json = parse_excel_to_schema_json(excel_file)
        # 输出格式化 JSON 到控制台
        print(json.dumps(schema_json, ensure_ascii=False, indent=4))

        # （可选）保存到文件
        # with open("schema_output.json", "w", encoding="utf-8") as f:
        #     json.dump(schema_json, f, ensure_ascii=False, indent=4)
        # print("✅ 已保存到 schema_output.json")

    except Exception as e:
        print(f"错误: {e}", file=sys.stderr)
