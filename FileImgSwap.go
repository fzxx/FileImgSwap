package main

import (
	"bufio"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"math"
	"os"
	"strings"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		return
	}

	inputPath := os.Args[1]
	
	if strings.HasSuffix(strings.ToLower(inputPath), ".fis.png") {
		restoreFromPNG(inputPath)
		return
	}
	
	convertToPNG(inputPath)
}

func printUsage() {
	fmt.Println("FileImgSwap 文图变 v0.1")
	fmt.Println("作者：风之暇想")
	fmt.Println("项目地址: https://github.com/fzxx/FileImgSwap")
	fmt.Println("====================================================")
	fmt.Println("  1. 文件转PNG: 将文件拖放至程序或输入命令: FileImgSwap 文件名")
	fmt.Println("     输出文件名为: 原文件名.fis.png")
	fmt.Println("  2. PNG还原文件: 将带.fis.png的文件拖放至程序或输入命令: FileImgSwap 文件名.fis.png")
	bufio.NewReader(os.Stdin).ReadString('\n')
}

func convertToPNG(inputPath string) {
	outputPath := inputPath + ".fis.png"

	data, err := os.ReadFile(inputPath)
	if err != nil {
		fmt.Println("读取文件错误:", err)
		return
	}
	dataLen := len(data)

	pixelCount := (dataLen + 7) / 8 + 1
	sideLength := int(math.Ceil(math.Sqrt(float64(pixelCount))))
	img := image.NewNRGBA64(image.Rect(0, 0, sideLength, sideLength))
	sizePixel := color.NRGBA64{
		R: uint16(dataLen & 0xFFFF),
		G: uint16((dataLen >> 16) & 0xFFFF),
		B: uint16((dataLen >> 32) & 0xFFFF),
		A: uint16((dataLen >> 48) & 0xFFFF),
	}
	img.SetNRGBA64(0, 0, sizePixel)
	for i := 1; i < pixelCount; i++ {
		x, y := i%sideLength, i/sideLength
		dataOffset := (i - 1) * 8
		r, g, b, a := uint16(0), uint16(0), uint16(0), uint16(0)
		if dataOffset < dataLen {
			r = uint16(data[dataOffset])
			if dataOffset+1 < dataLen {
				r |= uint16(data[dataOffset+1]) << 8
			}
		}
		if dataOffset+2 < dataLen {
			g = uint16(data[dataOffset+2])
			if dataOffset+3 < dataLen {
				g |= uint16(data[dataOffset+3]) << 8
			}
		}
		if dataOffset+4 < dataLen {
			b = uint16(data[dataOffset+4])
			if dataOffset+5 < dataLen {
				b |= uint16(data[dataOffset+5]) << 8
			}
		}
		if dataOffset+6 < dataLen {
			a = uint16(data[dataOffset+6])
			if dataOffset+7 < dataLen {
				a |= uint16(data[dataOffset+7]) << 8
			}
		}

		img.SetNRGBA64(x, y, color.NRGBA64{R: r, G: g, B: b, A: a})
	}

	f, err := os.Create(outputPath)
	if err != nil {
		fmt.Println("创建文件错误:", err)
		return
	}
	defer f.Close()

	encoder := png.Encoder{CompressionLevel: png.NoCompression}
	if err := encoder.Encode(f, img); err != nil {
		fmt.Println("编码PNG错误:", err)
		return
	}

	fmt.Println("转换完成!")
	fmt.Println("输入文件:", inputPath)
	fmt.Println("输入大小:", dataLen, "字节")
	fmt.Println("输出文件:", outputPath)
	fmt.Println("图片尺寸:", sideLength, "x", sideLength)
}

func restoreFromPNG(pngPath string) {
	originalPath := strings.TrimSuffix(pngPath, ".fis.png")
	
	if _, err := os.Stat(originalPath); err == nil {
		fmt.Println("错误: 原始文件", originalPath, "已存在，避免覆盖")
		return
	}

	file, err := os.Open(pngPath)
	if err != nil {
		fmt.Println("打开PNG文件错误:", err)
		return
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		fmt.Println("解码PNG错误:", err)
		return
	}

	nrgba64Img, ok := img.(*image.NRGBA64)
	if !ok {
		fmt.Println("错误: 该PNG文件不是16位RGBA格式，无法还原")
		return
	}

	bounds := img.Bounds()
	width, height := bounds.Max.X, bounds.Max.Y
	pixelCount := width * height
	sizePixel := nrgba64Img.NRGBA64At(0, 0)
	originalSize := int(sizePixel.R) | 
	                int(sizePixel.G)<<16 | 
	                int(sizePixel.B)<<32 | 
	                int(sizePixel.A)<<48
	data := make([]byte, 0, originalSize)
	for i := 1; i < pixelCount; i++ {
		if len(data) >= originalSize {
			break
		}
		
		x, y := i%width, i/width
		pixel := nrgba64Img.NRGBA64At(x, y)
		addBytes := func(b byte) {
			if len(data) < originalSize {
				data = append(data, b)
			}
		}
		
		addBytes(byte(pixel.R & 0xFF))
		addBytes(byte(pixel.R >> 8))
		addBytes(byte(pixel.G & 0xFF))
		addBytes(byte(pixel.G >> 8))
		addBytes(byte(pixel.B & 0xFF))
		addBytes(byte(pixel.B >> 8))
		addBytes(byte(pixel.A & 0xFF))
		addBytes(byte(pixel.A >> 8))
	}
	if len(data) > originalSize {
		data = data[:originalSize]
	}

	if err = os.WriteFile(originalPath, data, 0644); err != nil {
		fmt.Println("写入原始文件错误:", err)
		return
	}

	fmt.Println("还原完成!")
	fmt.Println("输入PNG:", pngPath)
	fmt.Println("还原文件:", originalPath)
	fmt.Println("原始大小:", originalSize, "字节")
	fmt.Println("还原大小:", len(data), "字节")
}
