# Go 图片缩放工具

[English](./README.md) | [繁體中文](./README.zh-tw.md) | [简体中文](./README.zh-cn.md)

[![Lint and Testing](https://github.com/appleboy/resize-tool/actions/workflows/testing.yml/badge.svg)](https://github.com/appleboy/resize-tool/actions/workflows/testing.yml)
[![Trivy Security Scan](https://github.com/appleboy/resize-tool/actions/workflows/trivy.yml/badge.svg)](https://github.com/appleboy/resize-tool/actions/workflows/trivy.yml)

[![幻灯片](https://img.shields.io/badge/幻灯片-SpeakerDeck-009287?logo=speakerdeck)](https://speakerdeck.com/appleboy/the-smart-choice-for-command-line-image-resizing)

![Go 图片缩放工具](./images/resize-tool.png)

一个用 Go 语言开发，简单但功能强大的图片缩放工具。

## 为什么需要这个工具？

在数字内容创作的时代，图片处理已经成为日常工作中不可或缺的一环。无论你是博主（Blogger）、社交媒体运营者、网页开发者，还是需要处理大量图片的内容创作者，都经常面临以下挑战：

**博主与内容创作者的痛点：**

- 撰写文章时需要将高分辨率照片缩小，以加快网页加载速度
- 不同平台（Facebook、Instagram、Medium）对图片尺寸有不同要求
- 批量处理旅游照片或产品图片时，使用在线工具既耗时又有隐私顾虑
- 图形界面软件（如 Photoshop）功能强大但过于庞大，只为了调整图片大小显得杀鸡用牛刀

**其他常见需求场景：**

- **网页开发者**：需要快速生成响应式图片的多种尺寸（thumbnail、medium、large）
- **电商从业者**：产品图片需要统一尺寸和文件大小，以维持网站一致性和性能
- **数字营销人员**：为不同广告平台准备符合规格的视觉素材
- **摄影师**：需要快速生成作品集缩略图或客户预览图
- **系统管理员**：需要在服务器端自动化处理用户上传的图片

这个命令行工具的设计初衷，就是提供一个**快速、轻量、可自动化**的解决方案。你不需要打开笨重的图形软件，不需要担心在线服务的隐私问题，只需要一行命令就能完成图片调整。更重要的是，它支持批量处理和脚本集成，让重复性的图片处理工作变得轻松简单，让创作者能把时间专注在真正重要的内容创作上。

## 目录

- [Go 图片缩放工具](#go-图片缩放工具)
  - [为什么需要这个工具？](#为什么需要这个工具)
  - [目录](#目录)
  - [功能特性](#功能特性)
  - [安装方式](#安装方式)
    - [通过脚本安装](#通过脚本安装)
      - [脚本自定义](#脚本自定义)
    - [从源码构建](#从源码构建)
    - [直接使用](#直接使用)
  - [使用说明](#使用说明)
    - [显示帮助](#显示帮助)
    - [基本用法](#基本用法)
    - [CLI 高级用法](#cli-高级用法)
  - [参数说明](#参数说明)
  - [输出文件名格式](#输出文件名格式)
  - [示例](#示例)
    - [1. 批量处理多张图片](#1-批量处理多张图片)
    - [2. 网站图片优化](#2-网站图片优化)
    - [3. 创建缩略图](#3-创建缩略图)
    - [4. 其他实用示例](#4-其他实用示例)
  - [支持的图片格式](#支持的图片格式)
  - [构建说明](#构建说明)
  - [性能提示](#性能提示)
  - [错误处理](#错误处理)
  - [技术细节](#技术细节)
    - [使用的库](#使用的库)
    - [图片处理算法](#图片处理算法)
  - [许可证](#许可证)

## 功能特性

- 支持多种图片格式：JPEG、PNG、GIF、TIFF、BMP
- **🎯 智能等比例缩放**：只指定宽度或高度时，另一边自动等比例计算
- **🎯 Glob 模式支持**：使用通配符模式如 `*.png` 或 `images/**/*.jpg` 处理多个文件
- 灵活的缩放选项
- 可选择是否保持宽高比
- 可调节 JPEG 质量
- 支持目录批量处理
- 支持并行处理提升效率
- 可自定义输出目录
- 详细显示进度和文件大小信息

## 安装方式

### 通过脚本安装

你可以使用提供的安装脚本，安装适用于你平台的最新预编译二进制文件：

```bash
curl -fsSL https://raw.githubusercontent.com/appleboy/resize-tool/main/install.sh | bash
```

或使用 wget：

```bash
wget -qO- https://raw.githubusercontent.com/appleboy/resize-tool/main/install.sh | bash
```

默认会安装到 `~/.resize-tool/bin/resize-tool`，并自动添加到 shell 的 PATH。

#### 脚本自定义

你可以通过设置环境变量自定义安装行为：

- `VERSION`：安装指定版本（默认最新发布版）
- `INSTALL_DIR`：更改安装目录（默认：`~/.resize-tool/bin`）
- `CURL_INSECURE=true`：允许不安全的 SSL 下载（不推荐）

示例：

```bash
INSTALL_DIR="$HOME/bin" VERSION="1.2.3" bash <(curl -fsSL https://raw.githubusercontent.com/appleboy/resize-tool/main/install.sh)
```

安装完成后，请重启终端或执行 `source ~/.bashrc`（或你的 shell 配置文件）以更新 PATH。

### 从源码构建

```bash
git clone <your-repo>
cd resize-tool
go mod tidy
go build -o resize-tool .
```

### 直接使用

如果你已经有编译好的二进制文件，可以直接使用：

```bash
resize-tool [选项] <图片文件>
```

## 使用说明

### 显示帮助

```bash
resize-tool --help
```

### 基本用法

```bash
# 默认宽度为 800 像素（高度自动等比例计算）
resize-tool image.jpg

# 🎯 只指定宽度，高度自动等比例计算（推荐）
resize-tool -w 1200 image.jpg

# 🎯 只指定高度，宽度自动等比例计算（推荐）
resize-tool --height 800 image.jpg

# 指定精确尺寸（可能会变形）
resize-tool -w 1200 --height 800 image.jpg

# 指定尺寸并保持宽高比（缩放至指定范围内）
resize-tool -k -w 1200 --height 800 image.jpg
```

### CLI 高级用法

```bash
# 设置 JPEG 质量（1-100）
resize-tool -q 85 -w 1000 image.jpg

# 指定输出目录
resize-tool -w 800 -o ./resized/ image.jpg

# 覆盖原始文件（不生成带尺寸的新文件名）
resize-tool -w 800 --overwrite image.jpg

# 🎯 使用 glob 模式处理多个文件
# 用引号包住 - 程序展开模式
resize-tool -w 1200 "images/*.png"

# 不用引号 - shell 展开模式（两种都可以）
resize-tool -w 1200 images/*.png

# 🎯 使用 glob 模式处理子目录中的文件
resize-tool -w 1920 "photos/**/*.jpg"

# 🎯 使用特定文件模式并指定输出目录
resize-tool -w 800 -o ./resized/ "vacation_*.jpg"

# 🎯 直接处理多个指定的文件
resize-tool -w 1024 file1.png file2.jpg file3.png

# 批量处理目录下所有图片
resize-tool -b -w 1200 /path/to/image/directory

# 批量处理并覆盖原始文件
resize-tool -b -w 1200 --overwrite /path/to/image/directory

# 批量处理时使用多线程
resize-tool -b --workers 8 -w 1920 /path/to/image/directory

# 详细输出模式
resize-tool -v -w 800 image.jpg

# 组合多个选项
resize-tool -w 1920 --height 1080 -q 90 -o ./output/ -k -v image.jpg
```

## 参数说明

| 参数           | 简写 | 默认值 | 说明                                 |
| -------------- | ---- | ------ | ------------------------------------ |
| `--width`      | `-w` | 0      | 输出宽度（像素，0=根据高度自动计算） |
| `--height`     |      | 0      | 输出高度（像素，0=根据宽度自动计算） |
| `--quality`    | `-q` | 95     | JPEG 质量（1-100）                   |
| `--output`     | `-o` | 同输入 | 输出目录（默认与输入相同）           |
| `--keep-ratio` | `-k` | false  | 同时指定宽高时，是否保持宽高比       |
| `--batch`      | `-b` | false  | 批量处理目录下所有图片               |
| `--workers`    |      | 4      | 批量处理时的并行线程数               |
| `--verbose`    | `-v` | false  | 启用详细输出                         |
| `--overwrite`  |      | false  | 覆盖原始文件，不创建新文件           |
| `--help`       | `-h` |        | 显示帮助信息                         |

## 输出文件名格式

缩放后的文件会自动包含尺寸信息：

- 原始文件：`photo.jpg`
- 输出文件：`photo_800x600.jpg`（实际高度取决于原图宽高比）

**注意**：使用 `--overwrite` 时，会直接替换原始文件，不会添加尺寸后缀。

## 示例

### 1. 批量处理多张图片

**使用 Glob 模式（推荐）：**

```bash
# 🎯 处理所有 PNG 文件（用引号 - 程序展开）
resize-tool -w 1200 "images/*.png"

# 🎯 处理所有 PNG 文件（不用引号 - shell 展开，两种都可以）
resize-tool -w 1200 images/*.png

# 🎯 处理当前目录下所有 JPG 文件
resize-tool --height 800 "*.jpg"

# 🎯 直接处理多个指定的文件
resize-tool -w 1920 -o ./resized/ file1.jpg file2.png file3.jpg

# 🎯 处理特定前缀的所有图片
resize-tool -w 1920 -o ./resized/ "vacation_*.{jpg,png}"

# 🎯 处理子目录中的图片（取决于 shell 支持）
resize-tool -w 1024 "photos/**/*.jpg"
```

**使用 Shell 循环：**

```bash
# 处理当前目录下所有 jpg 文件
for img in *.jpg; do
    ./resize-tool -w 1200 "$img"
done
```

```bash
# 处理当前目录下所有 png 文件（仅指定高度）
for img in *.png; do
    ./resize-tool --height 800 "$img"
done
```

### 2. 网站图片优化

```bash
# 创建三种不同尺寸（智能等比例缩放）
./resize-tool -w 1920 -q 85 -o ./large/ image.jpg
./resize-tool -w 1200 -q 85 -o ./medium/ image.jpg
./resize-tool -w 600 -q 80 -o ./small/ image.jpg
```

### 3. 创建缩略图

```bash
# 创建正方形缩略图（固定尺寸，可能会裁剪）
./resize-tool -w 300 --height 300 -o ./thumbnails/ image.jpg

# 创建缩略图（保持宽高比，最大 300x300）
./resize-tool -w 300 --height 300 -k -o ./thumbnails/ image.jpg
```

### 4. 其他实用示例

```bash
# 只指定宽度，高度自动计算
./resize-tool -w 1200 image.jpg

# 只指定高度，宽度自动计算
./resize-tool --height 800 image.jpg

# 同时指定宽高（可能会变形）
./resize-tool -w 1200 --height 800 image.jpg

# 同时指定宽高并保持宽高比（缩放至范围内）
./resize-tool -k -w 1200 --height 800 image.jpg

# 设置 JPEG 质量
./resize-tool -q 85 -w 1000 image.jpg

# 指定输出目录
./resize-tool -w 800 -o ./resized/ image.jpg

# 覆盖原始文件（直接替换）
./resize-tool -w 800 --overwrite image.jpg

# 🎯 使用 glob 模式处理多个文件
./resize-tool -w 1200 "images/*.png"
./resize-tool -w 1200 images/*.png  # 不用引号也可以
./resize-tool --height 800 "photos/*.jpg"

# 🎯 直接处理多个指定文件
./resize-tool -w 1920 -o ./output/ file1.jpg file2.png file3.jpg

# 批量处理目录下所有图片
./resize-tool -b -w 1200 /path/to/image/directory

# 批量处理并覆盖原始文件
./resize-tool -b -w 1200 --overwrite /path/to/image/directory

# 批量处理并使用多线程
./resize-tool -b --workers 8 -w 1920 /path/to/image/directory

# 详细输出模式
./resize-tool -v -w 800 image.jpg

# 组合多个选项（注意：--overwrite 不能与 --output 同时使用）
./resize-tool -w 1920 --height 1080 -q 90 -o ./output/ -k -v image.jpg
```

## 支持的图片格式

- **输入格式**：JPEG、PNG、GIF、TIFF、BMP
- **输出格式**：与输入格式相同

## 构建说明

如需从源码构建：

```bash
go build -o resize-tool .
```

跨平台构建（多平台）：

```bash
make release  # 构建多平台二进制文件
```

## 性能提示

- 使用 Lanczos 算法进行高质量图片缩放
- 处理大文件时可能需要更多内存
- JPEG 质量设置会影响文件大小和图片质量

## 错误处理

本工具会自动处理常见错误情况：

- 找不到文件
- 不支持的图片格式
- 输出目录创建失败
- 内存不足

## 技术细节

### 使用的库

- `github.com/disintegration/imaging` - 图片处理
- `github.com/spf13/cobra` - CLI 界面

### 图片处理算法

- **缩放算法**：Lanczos（高质量）
- **宽高比保持**：使用 Fit 方法，将图片缩放至指定范围内
- **强制尺寸**：使用 Resize 方法，可能会改变宽高比

## 许可证

MIT License
