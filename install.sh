#!/bin/sh
version=1.0.5
platform=linux

base="$(dirname "$(realpath "$0")")"
target_dir="$base/focas"

echo "Подготовка библиотеки в: $target_dir"

_arch=$(arch)
_set=true
if [ -z "${TARGETPLATFORM}" ]; then _set=false; fi

# Определение архитектуры
if [ "$_set" = true ] && [ "$TARGETPLATFORM" = "aarch64" ] || [ "$_arch" = "aarch64" ]; then
  arch=armv7
elif [ "$_set" = true ] && [ "$TARGETPLATFORM" = "linux/amd64" ] || [ "$_arch" = "x86_64" ]; then
  arch=x64
elif [ "$_set" = true ] && [ "$TARGETPLATFORM" = "linux/arm/v7" ] || [ "$_arch" = "armhf" ] || [ "$_arch" = "armv7l" ]; then
  arch=armv7
else
  arch=x86
fi

src_lib="libfwlib32-$platform-$arch.so.$version"

if [ ! -f "$base/$src_lib" ]; then
    echo "Ошибка: Файл '$src_lib' не найден в корне."
    exit 1
fi

mkdir -p "$target_dir"

# Копируем заголовок
cp "$base/fwlib32.h" "$target_dir/fwlib32.h"

# --- ИЗМЕНЕНИЕ ЗДЕСЬ ---
# Вместо ссылок делаем физические копии. 
# Это гарантирует работу в cross-platform среде (Windows/WSL/Linux)

echo "Создание физических копий библиотеки..."

# 1. Оригинальное имя (для истории/справки)
cp "$base/$src_lib" "$target_dir/libfwlib32.so.$version"

# 2. Имя с версией .1 (часто требуется загрузчиком runtime loader, если есть SONAME)
cp "$base/$src_lib" "$target_dir/libfwlib32.so.1"

# 3. Имя .so (требуется для компилятора go/gcc при флаге -lfwlib32)
cp "$base/$src_lib" "$target_dir/libfwlib32.so"

echo "Готово! Теперь это реальные файлы, а не ссылки."
ls -l "$target_dir/libfwlib32"*