# Go 圖片縮放工具

[English](./README.md) | [繁體中文](./README.zh-tw.md) | [簡體中文](./README.zh-cn.md)

[![Lint and Testing](https://github.com/appleboy/resize-tool/actions/workflows/testing.yml/badge.svg)](https://github.com/appleboy/resize-tool/actions/workflows/testing.yml)

一個用 Go 語言打造，簡單但功能強大的圖片縮放工具。

## 目錄

- [Go 圖片縮放工具](#go-圖片縮放工具)
  - [目錄](#目錄)
  - [功能特色](#功能特色)
  - [安裝方式](#安裝方式)
    - [使用腳本安裝](#使用腳本安裝)
      - [腳本自訂](#腳本自訂)
    - [從原始碼建置](#從原始碼建置)
    - [直接使用](#直接使用)
  - [使用說明](#使用說明)
    - [顯示說明](#顯示說明)
    - [基本用法](#基本用法)
    - [進階 CLI 用法](#進階-cli-用法)
  - [參數說明](#參數說明)
  - [輸出檔名格式](#輸出檔名格式)
  - [範例](#範例)
    - [1. 批次處理多張圖片](#1-批次處理多張圖片)
    - [2. 網站圖片最佳化](#2-網站圖片最佳化)
    - [3. 建立縮圖](#3-建立縮圖)
    - [4. 其他實用範例](#4-其他實用範例)
  - [支援的圖片格式](#支援的圖片格式)
  - [建置說明](#建置說明)
  - [效能提示](#效能提示)
  - [錯誤處理](#錯誤處理)
  - [技術細節](#技術細節)
    - [使用的函式庫](#使用的函式庫)
    - [圖片處理演算法](#圖片處理演算法)
  - [授權](#授權)

## 功能特色

- 支援多種圖片格式：JPEG、PNG、GIF、TIFF、BMP
- **🎯 智慧等比例縮放**：只指定寬度或高度時，另一邊自動等比例計算
- 彈性的縮放選項
- 可選擇是否保持長寬比
- 可調整 JPEG 品質
- 支援目錄批次處理
- 支援平行處理提升效率
- 可自訂輸出目錄
- 詳細顯示進度與檔案大小資訊

## 安裝方式

### 使用腳本安裝

你可以使用提供的安裝腳本，安裝對應平台的最新預編譯執行檔：

```bash
curl -fsSL https://raw.githubusercontent.com/appleboy/resize-tool/main/install.sh | bash
```

或使用 wget：

```bash
wget -qO- https://raw.githubusercontent.com/appleboy/resize-tool/main/install.sh | bash
```

預設會安裝到 `~/.resize-tool/bin/resize-tool`，並自動加入 shell 的 PATH。

#### 腳本自訂

你可以透過設定環境變數自訂安裝行為：

- `VERSION`：安裝指定版本（預設為最新發行版）
- `INSTALL_DIR`：變更安裝目錄（預設：`~/.resize-tool/bin`）
- `CURL_INSECURE=true`：允許不安全的 SSL 下載（不建議）

範例：

```bash
INSTALL_DIR="$HOME/bin" VERSION="1.2.3" bash <(curl -fsSL https://raw.githubusercontent.com/appleboy/resize-tool/main/install.sh)
```

安裝完成後，請重新啟動終端機或執行 `source ~/.bashrc`（或你的 shell 設定檔）以更新 PATH。

### 從原始碼建置

```bash
git clone <your-repo>
cd resize-tool
go mod tidy
go build -o resize-tool .
```

### 直接使用

如果你已經有編譯好的執行檔，可以直接使用：

```bash
resize-tool [選項] <圖片檔案>
```

## 使用說明

### 顯示說明

```bash
resize-tool --help
```

### 基本用法

```bash
# 預設縮放為 800x600 像素
resize-tool image.jpg

# 🎯 只指定寬度，高度自動等比例計算（推薦）
resize-tool -w 1200 image.jpg

# 🎯 只指定高度，寬度自動等比例計算（推薦）
resize-tool --height 800 image.jpg

# 指定確切尺寸（可能會變形）
resize-tool -w 1200 --height 800 image.jpg

# 指定尺寸並保持長寬比（縮放至指定範圍內）
resize-tool -k -w 1200 --height 800 image.jpg
```

### 進階 CLI 用法

```bash
# 設定 JPEG 品質（1-100）
resize-tool -q 85 -w 1000 image.jpg

# 指定輸出目錄
resize-tool -w 800 -o ./resized/ image.jpg

# 覆蓋原始檔案（不產生帶尺寸的新檔名）
resize-tool -w 800 --overwrite image.jpg

# 批次處理目錄下所有圖片
resize-tool -b -w 1200 /path/to/image/directory

# 批次處理並覆蓋原始檔案
resize-tool -b -w 1200 --overwrite /path/to/image/directory

# 批次處理時使用多執行緒
resize-tool -b --workers 8 -w 1920 /path/to/image/directory

# 詳細輸出模式
resize-tool -v -w 800 image.jpg

# 組合多個選項
resize-tool -w 1920 --height 1080 -q 90 -o ./output/ -k -v image.jpg
```

## 參數說明

| 參數           | 短參數 | 預設值 | 說明                               |
| -------------- | ------ | ------ | ---------------------------------- |
| `--width`      | `-w`   | 0      | 輸出寬度（像素，0=依高度自動計算） |
| `--height`     |        | 0      | 輸出高度（像素，0=依寬度自動計算） |
| `--quality`    | `-q`   | 95     | JPEG 品質（1-100）                 |
| `--output`     | `-o`   | 同輸入 | 輸出目錄（預設與輸入相同）         |
| `--keep-ratio` | `-k`   | false  | 同時指定寬高時，是否保持長寬比     |
| `--batch`      | `-b`   | false  | 批次處理目錄下所有圖片             |
| `--workers`    |        | 4      | 批次處理時的平行執行緒數量         |
| `--verbose`    | `-v`   | false  | 啟用詳細輸出                       |
| `--overwrite`  |        | false  | 覆蓋原始檔案，不建立新檔案         |
| `--help`       | `-h`   |        | 顯示說明                           |

## 輸出檔名格式

縮放後的檔案會自動加上尺寸資訊：

- 原始檔案：`photo.jpg`
- 輸出檔案：`photo_800x600.jpg`

**注意**：使用 `--overwrite` 時，會直接替換原始檔案，不會加上尺寸後綴。

## 範例

### 1. 批次處理多張圖片

```bash
# 處理目前目錄下所有 jpg 檔
for img in *.jpg; do
    ./resize-tool -w 1200 "$img"
done
```

```bash
# 處理目前目錄下所有 png 檔（僅指定高度）
for img in *.png; do
    ./resize-tool --height 800 "$img"
done
```

### 2. 網站圖片最佳化

```bash
# 建立三種不同尺寸（智慧等比例縮放）
./resize-tool -w 1920 -q 85 -o ./large/ image.jpg
./resize-tool -w 1200 -q 85 -o ./medium/ image.jpg
./resize-tool -w 600 -q 80 -o ./small/ image.jpg
```

### 3. 建立縮圖

```bash
# 建立正方形縮圖（固定尺寸，可能會裁切）
./resize-tool -w 300 --height 300 -o ./thumbnails/ image.jpg

# 建立縮圖（保持長寬比，最大 300x300）
./resize-tool -w 300 --height 300 -k -o ./thumbnails/ image.jpg
```

### 4. 其他實用範例

```bash
# 只指定寬度，高度自動計算
./resize-tool -w 1200 image.jpg

# 只指定高度，寬度自動計算
./resize-tool --height 800 image.jpg

# 同時指定寬高（可能會變形）
./resize-tool -w 1200 --height 800 image.jpg

# 同時指定寬高並保持長寬比（縮放至範圍內）
./resize-tool -k -w 1200 --height 800 image.jpg

# 設定 JPEG 品質
./resize-tool -q 85 -w 1000 image.jpg

# 指定輸出目錄
./resize-tool -w 800 -o ./resized/ image.jpg

# 覆蓋原始檔案（直接替換）
./resize-tool -w 800 --overwrite image.jpg

# 批次處理目錄下所有圖片
./resize-tool -b -w 1200 /path/to/image/directory

# 批次處理並覆蓋原始檔案
./resize-tool -b -w 1200 --overwrite /path/to/image/directory

# 批次處理並使用多執行緒
./resize-tool -b --workers 8 -w 1920 /path/to/image/directory

# 詳細輸出模式
./resize-tool -v -w 800 image.jpg

# 組合多個選項（注意：--overwrite 不能與 --output 同時使用）
./resize-tool -w 1920 --height 1080 -q 90 -o ./output/ -k -v image.jpg
```

## 支援的圖片格式

- **輸入格式**：JPEG、PNG、GIF、TIFF、BMP
- **輸出格式**：與輸入格式相同

## 建置說明

若要從原始碼建置：

```bash
go build -o resize-tool .
```

跨平台建置（多平台）：

```bash
make release  # 建立多平台執行檔
```

## 效能提示

- 採用 Lanczos 演算法進行高品質縮放
- 處理大型檔案時可能需要較多記憶體
- JPEG 品質設定會影響檔案大小與圖片品質

## 錯誤處理

本工具會自動處理常見錯誤情況：

- 找不到檔案
- 不支援的圖片格式
- 輸出目錄建立失敗
- 記憶體不足

## 技術細節

### 使用的函式庫

- `github.com/disintegration/imaging` - 圖片處理
- `github.com/spf13/cobra` - CLI 介面

### 圖片處理演算法

- **縮放演算法**：Lanczos（高品質）
- **長寬比保持**：使用 Fit 方法，將圖片縮放至指定範圍內
- **強制尺寸**：使用 Resize 方法，可能會改變長寬比

## 授權

MIT License
