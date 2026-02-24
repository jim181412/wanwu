#!/usr/bin/env python
# coding=utf-8
"""
@File:	mongo_utils.py
@Time:	2024/10/23 11:00:39
@Author:	wangj1075(wangj1075@chinaunicom.cn)
@Desc:	mongoDB操作工具类
"""
import json
import os
import sys
import datetime
from pymongo import MongoClient
from settings import MONGO_URL


def mongoConnect(db, colname):
    """
    兼容官方的mongodb://host1:port1,host2:port2/db/collection
    :param hoststring:
    :param db:
    :param colname:
    :return:
    """

    _client = MongoClient(MONGO_URL, 0, connectTimeoutMS=10000, serverSelectionTimeoutMS=30000)
    _col = _client[db][colname]
    return _col

if __name__ == "__main__":
    log_col = mongoConnect('rag','rag_user_logs')
    att_res = log_col.find({"id": "aa"}, {"_id": 0})
    for att_item in att_res:
        print(json.dumps(att_item))