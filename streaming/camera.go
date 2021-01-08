package streaming

import (
	"fmt"

	"gocv.io/x/gocv"
)

// cameraSession is a struct that allows capture and light postprocessing of cameras' images
type cameraSession struct {
	cams       []*gocv.VideoCapture // camera handles
	camCount   int                  // amount of cameras in session
	lastFrames []gocv.Mat           // pointers to the last frames (to free from memory)
	Denoising  bool                 // flag for the non-local means denoising algorithm

	dests []string // Streaming destinations
}

// GetFrames gets a single frame from each of the cameras in the camera session
func (cs *cameraSession) GetFrames() (frames []*gocv.Mat, err error) {
	// Free OpenCV allocated memory of the previous last frames
	for i := 0; i < cs.camCount; i++ {
		cs.lastFrames[i].Close()
	}

	// Create new frames and process them
	for i := 0; i < cs.camCount; i++ {
		frame := gocv.NewMat()
		if !cs.cams[0].Read(&frame) {
			err = fmt.Errorf("unexpected error while reading from camera %d", i)
			return
		} else if frame.Empty() {
			err = fmt.Errorf("retrieved an empty frame from camera camera %d", i)
			return
		}

		// preprocessing:
		if cs.Denoising {
			// TODO: No built in denoising function. Perhaps a slight blur could help
			// https://gocv.io/cvscope/
		}

		cs.lastFrames[i] = frame
		frames = append(frames, &frame)
	}

	return
}

// Stream starts streaming to the specified addresses
func (cs *cameraSession) Stream() (err error) {
	// If no sources or AddDestination is ran at the same time then return
	// an error

	return
}

// AddDestination adds a streaming destinations
func (cs *cameraSession) AddDestination(ip string, endpoint string) (err error) {
	// If the Stream was called then this should return an error

	// Check if there is a Parkind server running at the given address

	// Append to the destination list

	return
}

// Close closes the camera session
func (cs *cameraSession) Close() (err error) {
	// For each of the cameras
	for i := 0; i < cs.camCount; i++ {
		// close the handle
		err = cs.cams[i].Close()
		if err != nil {
			return
		}

		// free memory of the last frame
		err = cs.lastFrames[i].Close()
		if err != nil {
			return
		}
	}

	return
}

// NewCameraSession creates a new camera session
func NewCameraSession(camCount int) (cs cameraSession, err error) {
	// Set up pointer array for the last frames
	cs.cams = make([]*gocv.VideoCapture, camCount)
	cs.lastFrames = make([]gocv.Mat, camCount)

	// Set up the camera handles and place empty mats in the frames
	for i := 0; i < camCount; i++ {
		c, err := gocv.OpenVideoCapture(i)
		if err != nil {
			return cs, err
		}
		mat := gocv.NewMat()

		cs.cams[i] = c
		cs.lastFrames[i] = mat
	}

	cs.camCount = len(cs.cams)
	return
}
