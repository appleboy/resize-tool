package main

// Supported image file extensions.
const (
	extJPG  = ".jpg"
	extJPEG = ".jpeg"
	extPNG  = ".png"
	extGIF  = ".gif"
	extTIFF = ".tiff"
	extTIF  = ".tif"
	extBMP  = ".bmp"
)

// supportedImageExts is the set of image file extensions the tool accepts.
var supportedImageExts = map[string]bool{
	extJPG:  true,
	extJPEG: true,
	extPNG:  true,
	extGIF:  true,
	extTIFF: true,
	extTIF:  true,
	extBMP:  true,
}
