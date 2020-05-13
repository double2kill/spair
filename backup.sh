OUT_DIR=./backup #临时备份目录

DATE=`date +%Y_%m_%d` #获取当前系统时间

mkdir -p $OUT_DIR

cp spair.db "$OUT_DIR/spair.db_$DATE"
