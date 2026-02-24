#!/bin/bash
export PATH="/usr/local/Ascend/ascend-toolkit/latest/bin:/usr/local/Ascend/ascend-toolkit/latest/compiler/ccec_compiler/bin:/usr/local/Ascend/ascend-toolkit/latest/tools/ccec_compiler/bin:/root/miniconda3/bin:/root/miniconda3/condabin:/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin"

cd /model_extend/rag_core
bash start_run.sh
sleep 2
bash start_sse.sh
sleep 2
bash start_asyn_add.sh

sleep 2
# URL解析及入库-单条
bash start_url_single.sh
sleep 2

cd /model_extend/rag_es_server_unify
bash start_es_server.sh
sleep 2

cd /model_extend/minio_project
bash minio_start.sh
sleep 2

cd /model_extend/rag_core/graph
bash start_server.sh
sleep 2

# ================== hhh use ===================

sleep 99999d

