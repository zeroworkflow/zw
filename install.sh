#!/bin/bash
set -e

echo "Определение последней версии zeroworkflow/zw..."

LATEST_TAG=$(curl -s "https://api.github.com/repos/zeroworkflow/zw/releases/latest" | grep -Po '"tag_name": "\K.*?(?=")')

if [ -z "$LATEST_TAG" ]; then
  echo "Ошибка: не удалось найти тег последнего релиза. Проверьте подключение к сети."
  exit 1
fi

echo "Найдена последняя версия: $LATEST_TAG"

INSTALL_SCRIPT_URL="https://github.com/zeroworkflow/zw/releases/download/$LATEST_TAG/install.sh"

echo "Загрузка и выполнение установочного скрипта из релиза..."
echo "URL: $INSTALL_SCRIPT_URL"

curl -L "$INSTALL_SCRIPT_URL" | bash

echo "Установка последней версии zeroworkflow/zw завершена."