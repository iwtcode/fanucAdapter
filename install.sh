#!/bin/sh
version=1.0.5
platform=linux

# Определяем абсолютный путь к папке, где лежит скрипт (корень проекта)
base="$(dirname "$(realpath "$0")")"
# Целевая папка теперь focas внутри проекта
target_dir="$base/focas"

echo "Project root: $base"
echo "Target directory: $target_dir"

_arch=$(arch)
_set=true
if [ -z "${TARGETPLATFORM}" ]; then _set=false; fi

# 1. Определение архитектуры (оставляем логику выбора файла)
if [ "$_set" = true ] && [ "$TARGETPLATFORM" = "aarch64" ] || [ "$_arch" = "aarch64" ]; then
  # Для ARM (если нужно доустановить зависимости, раскомментируйте apt-get)
  # export LD_LIBRARY_PATH=$LD_LIBRARY_PATH:/usr/arm-linux-gnueabihf/lib/
  # dpkg --add-architecture armhf && apt-get update && apt-get install -y libc6-dev:armhf gcc-arm-linux-gnueabihf
  arch=armv7
elif [ "$_set" = true ] && [ "$TARGETPLATFORM" = "linux/amd64" ] || [ "$_arch" = "x86_64" ]; then
  arch=x64
elif [ "$_set" = true ] && [ "$TARGETPLATFORM" = "linux/arm/v7" ] || [ "$_arch" = "armhf" ] || [ "$_arch" = "armv7l" ]; then
  arch=armv7
else
  arch=x86
fi

echo "Detected architecture: $arch"

# Имя исходного файла библиотеки (например, libfwlib32-linux-x64.so.1.0.5)
src_lib="libfwlib32-$platform-$arch.so.$version"

# Проверяем наличие исходного файла
if [ ! -f "$base/$src_lib" ]; then
    echo "Error: Source file '$src_lib' not found in $base"
    exit 1
fi

# 2. Создаем папку focas, если её нет
mkdir -p "$target_dir"

# 3. Копируем библиотеку
echo "Copying library..."
cp "$base/$src_lib" "$target_dir/libfwlib32.so.$version"

# 4. Копируем заголовочный файл
echo "Copying header..."
cp "$base/fwlib32.h" "$target_dir/fwlib32.h"

# 5. Создаем симлинки внутри папки focas
# Переходим в папку, чтобы ссылки были относительными (правильными)
cd "$target_dir" || exit

# Удаляем старые ссылки, если есть
rm -f libfwlib32.so
rm -f libfwlib32.so.1

# Создаем новые (libfwlib32.so -> libfwlib32.so.1 -> libfwlib32.so.1.0.5)
ln -s "libfwlib32.so.$version" "libfwlib32.so.1"
ln -s "libfwlib32.so.1" "libfwlib32.so"

echo "Done!"
echo "Files in $target_dir:"
ls -l libfwlib32* fwlib32.h