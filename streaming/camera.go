package streaming

import (
	"bytes"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"sync"
	"time"

	"github.com/TheSlipper/ParkindStreamer/logging"
	"gocv.io/x/gocv"
)

// cameraSession is a struct that allows capture and light postprocessing of cameras' images
type cameraSession struct {
	devID int // id of the device that is running this session

	cams       []*gocv.VideoCapture // camera handles
	fps        int                  // cameras framerate
	camCount   int                  // amount of cameras in session
	lastFrames []gocv.Mat           // pointers to the last frames (to free from memory)
	Denoising  bool                 // flag for the non-local means denoising algorithm

	dests []string   // Streaming destinations
	smu   sync.Mutex // streaming (and destination management) mutex
	stop  bool       // stop streaming flag
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
	// Get the mutex
	cs.smu.Lock()
	defer cs.smu.Unlock()

	// If no sources or AddDestination then return an error
	if len(cs.dests) <= 0 {
		return errors.New("insufficient amount of streaming destinations")
	}

	// Stream images in the given framerate as long as the stop flag is not set to true
	recFrames := 0
	interval := time.Duration(int64(time.Second) / int64(cs.fps))
	start := time.Now()

	for !cs.stop {
		// If a second since start has passed reset the counters
		durUntilNow := time.Since(start)
		if durUntilNow >= time.Second {
			if recFrames < cs.fps {
				logging.WarningLog("dropped", strconv.Itoa(cs.fps-recFrames), "frames")
			}
			recFrames = 0
			start = time.Now()
			continue
		} else if recFrames == cs.fps {
			// if recorded enough frames in this second then just wait until the end of it
			time.Sleep(time.Second - durUntilNow)
			continue
		}

		// Pull the frames and increment the recorded frames
		_, err = cs.GetFrames()
		if err != nil {
			return err
		}
		recFrames++

		// Send the frames
		err = cs.send()
		if err != nil {
			return
		}

		// Wait for the calculated amount of time
		time.Sleep(interval)
	}

	cs.stop = false
	return
}

// AddDestination adds a streaming destinations
func (cs *cameraSession) AddDestination(ip string, endpoint string) (err error) {
	// Get the mutex
	cs.smu.Lock()
	defer cs.smu.Unlock()

	// Check if there is a Parkind server running at the given address
	// and if it will accept frames from this program instance
	// TODO

	// Append to the destination list
	path := fmt.Sprintf("http://%s/%s", ip, endpoint)
	cs.dests = append(cs.dests, path)

	return
}

// Close closes the camera session
func (cs *cameraSession) Close() (err error) {
	// Stop streaming
	cs.stop = true

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

// send sends all of the last recorded frames to specified destinations
func (cs *cameraSession) send() error {
	for i, frame := range cs.lastFrames {
		// Encode image
		data, err := gocv.IMEncode(".jpg", frame)
		if err != nil {
			return err
		}

		for _, dest := range cs.dests {
			// Send a POST request
			destFull := fmt.Sprintf("%s/%d/%d", dest, cs.devID, i)
			resp, err := http.Post(destFull, "image/jpeg", bytes.NewReader(data))
			if err != nil {
				return err
			}

			// Check if the image was read correctly
			if resp.StatusCode != 202 { // TODO: Confirm if it's this status code
				return fmt.Errorf("received incorrect http status code %d", resp.StatusCode)
			}
		}
	}
	return nil
}

// NewCameraSession creates a new camera session
func NewCameraSession(camCount int, fps int) (cs cameraSession, err error) {
	// If no cameras or invalid amount of cameras then stop
	if camCount <= 0 {
		return cs, fmt.Errorf("invalid amount of cameras %d", camCount)
	}

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

	// Set up other variables
	cs.fps = fps
	cs.camCount = len(cs.cams)
	cs.stop = false // do not order the stream to stop

	return
}
