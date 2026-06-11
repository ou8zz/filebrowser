#!/bin/sh

set -e

# Ensure configuration exists
if [ ! -f "/config/settings.json" ]; then
  cp -a /defaults/settings.json /config/settings.json
fi

# Extract config file path from arguments
config_file=""
next_is_config=0
for arg in "$@"; do
  if [ "$next_is_config" -eq 1 ]; then
    config_file="$arg"
    break
  fi
  case "$arg" in
    -c|--config)
      next_is_config=1
      ;;
    -c=*|--config=*)
      config_file="${arg#*=}"
      break
      ;;
  esac
done

# If no config argument is provided, set the default and add it to the args                                                                 
if [ -z "$config_file" ]; then 
  config_file="/config/settings.json"                                                                                                                                                                                                 
  set -- --config=/config/settings.json "$@"                                                                                                       
fi                                                                                                                                                                                                                                                                                                                                                             

# =========================================================
# 【新增：UMASK 权限解析与应用】
# =========================================================
if [ -n "$UMASK" ]; then
  echo "Applying UMASK=${UMASK}"
  # 校验传入的是否是合法的 3 位或 4 位八进制数字（防止用户手抖写错引发系统故障）
  if echo "$UMASK" | grep -Eq '^[0-7]{3,4}$'; then
    umask "$UMASK"
  else
    echo "Warning: Invalid UMASK format '$UMASK', falling back to 022."
    umask 022
  fi
else
  # 如果用户没有传入 UMASK 环境变量，则使用默认的 022
  umask 022
fi
# =========================================================

exec filebrowser "$@"
