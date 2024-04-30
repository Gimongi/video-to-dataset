package utils

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	data "gimongi/video-to-dataset/data"
)

// Extracts individual .jpg frames from a provided video with interval as the time between each frame
func ExtractVideoFrames(outputDir, inputFile string, interval float32) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil { // ensure outputDir exists and if not create it
		return fmt.Errorf("os.MkdirAll: %w", err)
	}

	duration, err := GetVideoDuration(inputFile)
	if err != nil {
		return fmt.Errorf("getVideoDuration: %w", err)
	}

	frameCount := int(duration / interval)
	if frameCount == 0 {
		return nil
	}
	for i := 0; i < frameCount; i++ {
		time := float32(i) * interval
		err := ExtractFrame(outputDir, inputFile, i, time)
		if err != nil {
			return fmt.Errorf("extractFrame: file: %s, time: %f, seconds: %w", inputFile, time, err)
		}
	}

	return nil
}

func ExtractFrame(outputDir, inputFile string, frameNum int, timestamp float32) error {
	outputFile := fmt.Sprintf("%sframe_%d.jpg", outputDir, frameNum)

	cmd := exec.Command(
		"ffmpeg",
		"-ss", strconv.FormatFloat(float64(timestamp), // specifies the start time offset
			'f', // format: fixed-point notation
			-1,  // precision: smallest number of digits necessary
			32), // float64 should be converted to a string as if it were a float32
		"-i", inputFile, // input file
		"-frames:v", "1", // number of frames to process
		outputFile, // output file
	)

	if err := cmd.Run(); err != nil {
		return fmt.Errorf("cmd.Run: %w", err)
	}

	return nil
}

/*******************
**** VIDEO INFO ****
*******************/

// Extracts width and height information about a video file using FFprobe
func GetVideoDimensions(inputVideo string) (width int, height int, err error) {
	// Command format: ffprobe -v error -select_streams v:0 -show_entries stream=width,height -of csv=s=x:p=0 <input video>
	cmd := exec.Command(
		"ffprobe",
		"-v", "error", // set log level to error to suppress non-error messages
		"-select_streams",     // limit output to certain streams (in this case, the video stream)
		"v:0",                 // select first video stream
		"-show_entries",       // specify which stream properties to display
		"stream=width,height", // display width and height of the video stream
		"-of",                 // specify output format
		"csv=s=x:p=0",         // output as comma-separated values in the format widthxheight
		inputVideo,            // input video file to analyze
	)

	output, err := cmd.Output()
	if err != nil {
		// return error
		log.Fatalf("FFprobe command failed for input video: %s\n%s", inputVideo, output)
	}

	s := strings.Split(strings.TrimSpace(string(output)), "x")
	wid, err := strconv.Atoi(s[0])
	if err != nil {
		return 0, 0, err
	}
	hei, err := strconv.Atoi(s[1])
	if err != nil {
		return 0, 0, err
	}

	fmt.Printf("Successfully extracted video dimensions. Width: %d, Height: %d", wid, hei)
	return wid, hei, nil
}

// Returns the duration of a video as a float. Ex: A 2.5 second vid: 2.50000...
func GetVideoDuration(fileName string) (float32, error) {
	// --Inform: specifies type of information
	// General: retrieve information about the "General" category
	// Duration: specifies we want to output duration info
	cmd := exec.Command(
		"mediainfo",
		"--Inform=General;%Duration%",
		fileName, // input file
	)

	output, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("cmd.Output: %v", err)
	}

	intMiliseconds, err := strconv.Atoi(strings.TrimSpace(string(output)))
	if err != nil {
		return 0, fmt.Errorf("strconv.Atoi: %v", err)
	}

	return float32(intMiliseconds) / 1000.0, nil
}

/**************************
**** OBJECT EXTRACTION ****
**************************/

// TODO: Specify output location? Or should it be auto-chosen
// Extract a single image from video
func ExtractVideoObject(file string, frameInfo *data.FrameInfo) {
	outputImage := fmt.Sprintf("image_%s_%f.png", frameInfo.TimeFrame, frameInfo.Size)

	// Command format: ffmpeg -ss <timeFrame> -i <video uri> -frames:v 1 -filter:v crop=<input> <output image>
	cmd := exec.Command(
		"ffmpeg",
		"-ss", frameInfo.TimeFrame, // seek to the specified time frame
		"-i", file, // input video file
		"-frames:v", "1", // extract only one video frame
		"-filter:v", "crop="+frameInfo.BoundingBox, // crop the frame to specified bounding box
		outputImage, // output extracted image to file
	)

	output, err := cmd.Output()
	if err != nil {
		log.Fatalf("FFmpeg command failed for time frame %s and bounding box %s: %v\n%s", frameInfo.TimeFrame, frameInfo.BoundingBox, err, string(output))
	}

	log.Printf("Image frame successfully extracted and cropped for time frame %s and bounding box %s.\n", frameInfo.TimeFrame, frameInfo.BoundingBox)
}

// TODO: Specify output location? Or should it be auto-chosen
// Extract a list of images from video
func ExtractVideoObjects(file string, frameInfos []*data.FrameInfo) {
	for i, frameInfo := range frameInfos {
		// TODO: name needs to clarify what this refers to more clearly (maybe also store in GCP)
		outputImage := fmt.Sprintf("image_%s_%f_%d.png", frameInfo.TimeFrame, frameInfo.Size, i+1)

		// Command format: ffmpeg -ss <timeFrame> -i <video uri> -frames:v 1 -filter:v crop=<input> <output image>
		cmd := exec.Command(
			"ffmpeg",
			"-ss", frameInfo.TimeFrame, // seek to the specified time frame
			"-i", file, // input video file
			"-frames:v", "1", // extract only one video frame
			"-filter:v", "crop="+frameInfo.BoundingBox, // crop the frame to specified bounding box
			outputImage, // output extracted image to file
		)

		// Run FFmpeg command
		output, err := cmd.CombinedOutput()
		if err != nil {
			log.Fatalf("FFmpeg command failed for time frame %s and bounding box %s: %v\n%s", frameInfo.TimeFrame, frameInfo.BoundingBox, err, string(output))
		}

		log.Printf("Image frame successfully extracted and cropped for time frame %s and bounding box %s.\n", frameInfo.TimeFrame, frameInfo.BoundingBox)
	}
}
