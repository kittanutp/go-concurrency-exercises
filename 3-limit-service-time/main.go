//////////////////////////////////////////////////////////////////////
//
// Your video processing service has a freemium model. Everyone has 10
// sec of free processing time on your service. After that, the
// service will kill your process, unless you are a paid premium user.
//
// Beginner Level: 10s max per request
// Advanced Level: 10s max per user (accumulated)
//

package main

import "time"

// User defines the UserModel. Use this to check whether a User is a
// Premium user or not
type User struct {
	ID        int
	IsPremium bool
	TimeUsed  int64 // in seconds
}

// HandleRequest runs the processes requested by users. Returns false
// if process had to be killed
func HandleRequest(process func(), u *User) bool {
	if !u.IsPremium && u.TimeUsed >= 10 {
		return false
	}
	done := make(chan bool)
	startTime := time.Now()
	go func() {
		process()
		done <- true

	}()

	timeRemaining := int64(10) - u.TimeUsed
	if timeRemaining < 0 {
		timeRemaining = 0
	}

	for {
		select {
		case <-done:
			timeTaken := time.Since(startTime).Seconds()
			if !u.IsPremium {
				u.TimeUsed += int64(timeTaken)
			}
			return true
		case <-time.After(time.Second * time.Duration(timeRemaining)):
			if !u.IsPremium {
				timeTaken := time.Since(startTime).Seconds()
				u.TimeUsed += int64(timeTaken)
				return false
			}

		}
	}
}

func main() {
	RunMockServer()
}
