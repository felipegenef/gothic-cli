package cli

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/nfnt/resize"
	webp "golang.org/x/image/webp"
)

type ImgOptimizationCommand struct {
	cli *GothicCli
}

func NewImgOptimizationCommandCli() ImgOptimizationCommand {
	return ImgOptimizationCommand{}
}

func (command *ImgOptimizationCommand) OptimizeImages() {
	command.setup()
	// TODO change it to struct properties
	inputDir := "./optimize"
	outputDir := "./public"
	// TODO change it to struct properties and make a for loop over slice
	downloadResizeCMD := exec.Command("go", "mod", "download", "github.com/nfnt/resize")
	downloadWebpCMD := exec.Command("go", "mod", "download", "golang.org/x/image")
	// Make sure needed packages have been downloaded
	if err := downloadResizeCMD.Run(); err != nil {
		log.Fatalf("Error executing add command: %v", err)
	}
	if err := downloadWebpCMD.Run(); err != nil {
		log.Fatalf("Error executing add command: %v", err)
	}

	config := command.cli.GetConfig()

	// Create the output directory if it doesn't exist
	if err := os.MkdirAll(outputDir, os.ModePerm); err != nil {
		panic(err)
	}

	// Read the files from the input directory
	files, err := os.ReadDir(inputDir)
	if err != nil {
		panic(err)
	}

	for _, file := range files {
		if !file.IsDir() {
			inputPath := filepath.Join(inputDir, file.Name())
			baseName := file.Name()[:len(file.Name())-len(filepath.Ext(file.Name()))]
			imageDir := filepath.Join(outputDir, baseName)
			ext := filepath.Ext(file.Name())

			if err := os.MkdirAll(imageDir, os.ModePerm); err != nil {
				fmt.Printf("Error creating directory %s: %v\n", imageDir, err)
				continue
			}

			// Open the image file
			imgFile, err := os.Open(inputPath)
			if err != nil {
				fmt.Printf("Error opening file %s: %v\n", inputPath, err)
				continue
			}

			// Decode the image
			var img image.Image
			switch ext {
			case ".png":
				img, err = png.Decode(imgFile)
			case ".jpg", ".jpeg":
				img, err = jpeg.Decode(imgFile)
			case ".webp":
				img, err = webp.Decode(imgFile)
			default:
				fmt.Printf("Unsupported file format: %s\n", ext)
				imgFile.Close()
				continue
			}
			imgFile.Close()
			if err != nil {
				fmt.Printf("Error decoding image %s: %v\n", inputPath, err)
				continue
			}

			// Get original dimensions
			originalWidth := uint(img.Bounds().Dx())
			originalHeight := uint(img.Bounds().Dy())

			// Calculate new dimensions for blurred image (20% of original)
			// Check if Deploy section exists and get lowResolutionRate
			lowResolutionRate := 20 // default value
			if config.OptimizeImages.LowResolutionRate > 0 {
				lowResolutionRate = config.OptimizeImages.LowResolutionRate
			}

			newWidth := originalWidth * uint(lowResolutionRate) / 100
			newHeight := originalHeight * uint(lowResolutionRate) / 100

			// Save the original image
			originalPath := filepath.Join(imageDir, "original"+ext)
			originalFile, err := os.Create(originalPath)
			if err != nil {
				fmt.Printf("Error creating original file %s: %v\n", originalPath, err)
				continue
			}
			defer originalFile.Close()

			// Resize the image to 20% of its original dimensions
			resizedImg := resize.Resize(newWidth, newHeight, img, resize.Lanczos3)

			// Save the resized image as blurred image
			blurredPath := filepath.Join(imageDir, "blurred"+ext)
			blurredFile, err := os.Create(blurredPath)
			if err != nil {
				fmt.Printf("Error creating blurred file %s: %v\n", blurredPath, err)
				continue
			}
			defer blurredFile.Close()

			switch ext {
			case ".png":
				if err := png.Encode(originalFile, img); err != nil {
					fmt.Printf("Error saving original PNG image %s: %v\n", originalPath, err)
				}
				if err := png.Encode(blurredFile, resizedImg); err != nil {
					fmt.Printf("Error saving blurred PNG image %s: %v\n", originalPath, err)
				}

			case ".jpg", ".jpeg":
				if err := jpeg.Encode(originalFile, img, &jpeg.Options{Quality: 100}); err != nil {
					fmt.Printf("Error saving original JPEG image %s: %v\n", originalPath, err)
				}
				if err := jpeg.Encode(blurredFile, resizedImg, &jpeg.Options{Quality: 20}); err != nil {
					fmt.Printf("Error saving blurred image %s: %v\n", blurredPath, err)
				}
			case ".webp":
				if err := png.Encode(originalFile, img); err != nil {
					fmt.Printf("Error saving original WebP image %s: %v\n", originalPath, err)
				}
				if err := png.Encode(blurredFile, resizedImg); err != nil {
					fmt.Printf("Error saving blurred WebP image %s: %v\n", originalPath, err)
				}
			}
		} else {
			fmt.Println("The 'optimizeImages' key was not found in gothic-config.json.")
		}

	}

	fmt.Println("Resizing complete!")
}

func (command *ImgOptimizationCommand) setup() {
	// TODO change it to struct properties and make a for loop over slice
	getResizeCMD := exec.Command("go", "get", "github.com/nfnt/resize")
	getWebpCMD := exec.Command("go", "get", "golang.org/x/image")
	// Make sure needed packages have been added to go.mod
	if err := getResizeCMD.Run(); err != nil {
		log.Fatalf("Error executing add command: %v", err)
	}
	if err := getWebpCMD.Run(); err != nil {
		log.Fatalf("Error executing add command: %v", err)
	}
	// TODO change it to struct properties and make a for loop over slice
	downloadResizeCMD := exec.Command("go", "mod", "download", "github.com/nfnt/resize")
	downloadWebpCMD := exec.Command("go", "mod", "download", "golang.org/x/image")
	// Make sure needed packages have been downloaded
	if err := downloadResizeCMD.Run(); err != nil {
		log.Fatalf("Error executing add command: %v", err)
	}
	if err := downloadWebpCMD.Run(); err != nil {
		log.Fatalf("Error executing add command: %v", err)
	}
}
