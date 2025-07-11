#!/bin/bash

# Go Image Resize Tool 使用範例腳本

echo "=== Go Image Resize Tool 使用範例 ==="
echo

# 顯示工具版本資訊
echo "1. 顯示幫助資訊："
./resize-tool --help
echo

echo "=== 基本使用範例 ==="
echo "注意：以下命令需要您提供實際的圖片檔案"
echo

echo "2. 基本調整大小（調整到 800x600）："
echo "   ./resize-tool image.jpg"
echo

echo "3. 只指定寬度（高度自動按比例計算）：⭐ NEW!"
echo "   ./resize-tool -w 1200 image.jpg"
echo

echo "4. 只指定高度（寬度自動按比例計算）：⭐ NEW!"
echo "   ./resize-tool --height 800 image.jpg"
echo

echo "5. 調整到指定尺寸（可能變形）："
echo "   ./resize-tool -w 1200 --height 800 image.jpg"
echo

echo "6. 調整到指定尺寸但保持比例（會縮放到框內）："
echo "   ./resize-tool -k -w 1200 --height 800 image.jpg"
echo

echo "5. 設定 JPEG 品質："
echo "   ./resize-tool -q 85 -w 1000 image.jpg"
echo

echo "6. 指定輸出目錄："
echo "   ./resize-tool -w 800 -o ./resized/ image.jpg"
echo

echo "7. 批次處理目錄中的所有圖片："
echo "   ./resize-tool -b -w 1200 /path/to/image/directory"
echo

echo "8. 批次處理並使用多執行緒："
echo "   ./resize-tool -b --workers 8 -w 1920 /path/to/image/directory"
echo

echo "9. 詳細輸出模式："
echo "   ./resize-tool -v -w 800 image.jpg"
echo

echo "10. 組合多個選項："
echo "    ./resize-tool -w 1920 --height 1080 -q 90 -o ./output/ -k -v image.jpg"
echo

echo "=== 實用腳本範例 ==="
echo

echo "⭐ 智慧調整：只設定寬度，高度自動計算："
echo "for img in *.jpg; do"
echo "    ./resize-tool -w 1200 \"\$img\""
echo "done"
echo

echo "⭐ 智慧調整：只設定高度，寬度自動計算："
echo "for img in *.png; do"
echo "    ./resize-tool --height 800 \"\$img\""
echo "done"
echo

echo "為網站建立三種不同尺寸（智慧比例）："
echo "./resize-tool -w 1920 -q 85 -o ./large/ image.jpg"
echo "./resize-tool -w 1200 -q 85 -o ./medium/ image.jpg"
echo "./resize-tool -w 600 -q 80 -o ./small/ image.jpg"
echo

echo "建立正方形縮圖（固定尺寸，可能裁切）："
echo "./resize-tool -w 300 --height 300 -o ./thumbnails/ image.jpg"
echo

echo "建立縮圖（保持比例，最大 300x300）："
echo "./resize-tool -w 300 --height 300 -k -o ./thumbnails/ image.jpg"
echo

echo "=== 支援的圖片格式 ==="
echo "輸入格式：JPEG, PNG, GIF, TIFF, BMP"
echo "輸出格式：與輸入格式相同"
echo

echo "=== 工具編譯 ==="
echo "如需重新編譯："
echo "go build -o resize-tool ."
echo
echo "如需交叉編譯："
echo "make release  # 建立多平台版本"
