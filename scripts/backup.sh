#!/usr/bin/env bash
# 数学协会网站 MySQL 每日备份脚本（在部署服务器上运行）
#
# 用法：
#   1. 将本脚本放到 /root/nzdmath/scripts/backup.sh
#   2. chmod +x backup.sh
#   3. 配置 crontab 每日凌晨 3 点执行：
#      crontab -e
#      0 3 * * * /root/nzdmath/scripts/backup.sh >> /root/backups/backup.log 2>&1
#
# 保留策略：保留最近 14 份备份，自动清理更早的。

set -euo pipefail

BACKUP_DIR="${BACKUP_DIR:-/root/backups}"
RETENTION_DAYS="${RETENTION_DAYS:-14}"
CONTAINER_NAME="${MYSQL_CONTAINER:-math-top-mysql}"
DB_NAME="${MYSQL_DB:-math_top}"

mkdir -p "$BACKUP_DIR"

TIMESTAMP=$(date +%F_%H%M%S)
OUT_FILE="$BACKUP_DIR/${DB_NAME}_${TIMESTAMP}.sql.gz"

# 密码从容器环境变量读取，不落盘、不出现在进程参数中
docker exec "$CONTAINER_NAME" sh -c 'mysqldump -uroot -p"$MYSQL_ROOT_PASSWORD" --single-transaction --quick --routines --triggers '"$DB_NAME" \
  | gzip > "$OUT_FILE"

# 校验备份非空，防止"成功输出空文件"的假备份
if [ ! -s "$OUT_FILE" ]; then
  echo "[$(date '+%F %T')] 备份失败：输出文件为空" >&2
  rm -f "$OUT_FILE"
  exit 1
fi

# 清理超过保留天数的旧备份
find "$BACKUP_DIR" -name "${DB_NAME}_*.sql.gz" -mtime "+$RETENTION_DAYS" -delete

# 上传文件目录（头像/作品/聊天文件）与数据库同等重要，一并打包
UPLOADS_DIR="${UPLOADS_DIR:-/root/nzdmath/backend/storage/uploads}"
if [ -d "$UPLOADS_DIR" ]; then
  tar -czf "$BACKUP_DIR/uploads_${TIMESTAMP}.tar.gz" -C "$UPLOADS_DIR" .
else
  echo "[$(date '+%F %T')] 警告：上传目录不存在，跳过备份：$UPLOADS_DIR" >&2
fi
find "$BACKUP_DIR" -name "uploads_*.tar.gz" -mtime "+$RETENTION_DAYS" -delete

echo "[$(date '+%F %T')] 备份成功：$OUT_FILE ($(du -h "$OUT_FILE" | cut -f1))"

# 恢复演练（可选，首次部署时执行一次验证流程可用）：
#   gunzip < 备份文件.sql.gz | docker exec -i math-top-mysql sh -c 'mysql -uroot -p"$MYSQL_ROOT_PASSWORD" math_top_restore_test'
