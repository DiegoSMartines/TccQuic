package test_client

import (
	"log"
	"sync"
	"time"
)

type PlaybackSimulator struct {
	currentPlaybackSegment int
	segmentPlaybackTime    []time.Time
	mutex                  *sync.Mutex
	cond                   *sync.Cond

	segmentDuration time.Duration
	baseLatency     time.Duration
	firstSegment    int
	lastSegment     int
}

func NewPlaybackSimulator(
	segmentDuration time.Duration,
	baseLatency time.Duration,
	firstSegment int,
	lastSegment int,
) *PlaybackSimulator {
	mutex := new(sync.Mutex)
	cond := sync.NewCond(mutex)
	return &PlaybackSimulator{
		currentPlaybackSegment: firstSegment - 1,
		segmentPlaybackTime:    make([]time.Time, lastSegment+1),
		mutex:                  mutex,
		cond:                   cond,

		segmentDuration: segmentDuration,
		baseLatency:     baseLatency,
		firstSegment:    firstSegment,
		lastSegment:     lastSegment,
	}
}

func (p *PlaybackSimulator) Start() {
	p.mutex.Lock()
	t := time.Now().Add(p.baseLatency)
	for i := p.firstSegment; i <= p.lastSegment; i++ {
		p.segmentPlaybackTime[i] = t
		t = t.Add(p.segmentDuration)
	}
	p.mutex.Unlock()

	go func() {
		for playbackSegment := p.firstSegment; playbackSegment <= p.lastSegment; playbackSegment++ {
			time.Sleep(time.Until(p.segmentPlaybackTime[playbackSegment]))

			log.Printf("Playback %d", playbackSegment)

			p.mutex.Lock()
			p.currentPlaybackSegment = playbackSegment
			p.mutex.Unlock()
			p.cond.Broadcast()
		}
	}()
}

func (p *PlaybackSimulator) WaitForPlaybackStart(segment int) {
	p.mutex.Lock()
	for p.currentPlaybackSegment < segment {
		p.cond.Wait()
	}
	p.mutex.Unlock()
}

// Returns the time remanining until segment starts being played.
//
// Returns 0 if the segment was already played or if it is not currently in
// the buffer.
func (p *PlaybackSimulator) GetTimeToReceive(segment int) time.Duration {
	var result time.Duration = 0

	p.mutex.Lock()
	if segment == p.currentPlaybackSegment+1 {
		result = time.Until(p.segmentPlaybackTime[segment])
	}
	p.mutex.Unlock()

	if result < 0 {
		return 0
	} else {
		return result
	}
}
