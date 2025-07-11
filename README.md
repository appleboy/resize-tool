# Go Image Resize Tool

一個簡單但功能強大的圖片調整大小工具，使用 Go 語言開發。

## 功能特色

- 支援多種圖片格式：JPEG, PNG, GIF, TIFF, BMP
- **🎯 智慧比例調整**：只指定寬度或高度時，另一維度會自動按比例計算
- 靈活的尺寸調整選項
- 保持長寬比例選項
- 可調整 JPEG 品質
- 批次處理目錄中的所有圖片
- 並行處理提高效率
- 自訂輸出目錄
- 詳細的進度和大小資訊顯示

## 安裝

### 從原始碼編譯

```bash
git clone <your-repo>
cd resize-tool
go mod tidy
go build -o resize-tool .
```

### 直接使用

如果你已經有編譯好的執行檔，可以直接使用：

```bash
./resize-tool [選項] <圖片檔案>
```

## 使用方法

### 基本用法

```bash
# 預設調整圖片到 800x600 像素
./resize-tool image.jpg

# 🎯 只指定寬度，高度自動按比例計算（推薦）
./resize-tool -w 1200 image.jpg

# 🎯 只指定高度，寬度自動按比例計算（推薦）
./resize-tool --height 800 image.jpg

# 指定具體尺寸（可能會變形）
./resize-tool -w 1200 --height 800 image.jpg

# 指定尺寸範圍但保持比例（縮放到框內）
./resize-tool -k -w 1200 --height 800 image.jpg
```

### 進階選項

```bash
# 設定 JPEG 品質（1-100）
./resize-tool -q 85 -w 1000 image.jpg

# 指定輸出目錄
./resize-tool -w 800 -o ./resized/ image.jpg

# 批次處理目錄中的所有圖片
./resize-tool -b -w 1200 /path/to/images/

# 使用多執行緒進行批次處理
./resize-tool -b --workers 8 -w 1920 /path/to/images/

# 詳細輸出模式
./resize-tool -v -w 800 image.jpg

# 組合多個選項
./resize-tool -w 1920 --height 1080 -q 90 -o ./output/ -k -v image.jpg
```

## 參數說明

| 參數           | 簡寫 | 預設值 | 說明                                     |
| -------------- | ---- | ------ | ---------------------------------------- |
| `--width`      | `-w` | 0      | 輸出寬度（像素，0=根據高度自動計算）     |
| `--height`     |      | 0      | 輸出高度（像素，0=根據寬度自動計算）     |
| `--quality`    | `-q` | 95     | JPEG 品質（1-100）                       |
| `--output`     | `-o` | 原目錄 | 輸出目錄                                 |
| `--keep-ratio` | `-k` | false  | 當同時指定寬高時保持長寬比例             |
| `--batch`      | `-b` | false  | 批次處理目錄中的所有圖片                 |
| `--workers`    |      | 4      | 批次處理時的並行執行緒數                 |
| `--verbose`    | `-v` | false  | 顯示詳細輸出                             |
| `--help`       | `-h` |        | 顯示幫助訊息                             |

## 輸出檔名格式

調整後的檔案會自動加上尺寸資訊：

- 原檔案：`photo.jpg`
- 輸出檔案：`photo_800x600.jpg`

## 範例

### 1. 批次處理多張圖片

```bash
# 對當前目錄所有 jpg 檔案進行調整
for img in *.jpg; do
    ./resize-tool -w 1200 --height 800 -k "$img"
done
```

### 2. 為網站最佳化圖片

```bash
# 建立三種不同尺寸
./resize-tool -w 1920 --height 1080 -q 85 -o ./large/ image.jpg
./resize-tool -w 1200 --height 800 -q 85 -o ./medium/ image.jpg
./resize-tool -w 600 --height 400 -q 80 -o ./small/ image.jpg
```

### 3. 保持比例的縮圖

```bash
# 建立不超過 300x300 的縮圖，保持原比例
./resize-tool -w 300 --height 300 -k -o ./thumbnails/ image.jpg
```

## 支援的圖片格式

- **輸入格式**：JPEG, PNG, GIF, TIFF, BMP
- **輸出格式**：與輸入格式相同

## 效能提示

- 使用 Lanczos 演算法進行高品質的圖片調整
- 大檔案處理可能需要較多記憶體
- JPEG 品質設定會影響檔案大小和圖片品質

## 錯誤處理

工具會自動處理常見的錯誤情況：

- 檔案不存在
- 不支援的圖片格式
- 輸出目錄建立失敗
- 記憶體不足

## 技術細節

### 使用的套件

- `github.com/disintegration/imaging` - 圖片處理
- `github.com/spf13/cobra` - CLI 介面

### 圖片處理演算法

- **調整演算法**：Lanczos（高品質）
- **保持比例**：使用 Fit 方法，圖片會縮放到指定尺寸內
- **強制尺寸**：使用 Resize 方法，可能會改變長寬比

## 授權

MIT License
