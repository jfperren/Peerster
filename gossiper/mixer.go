package gossiper

import (
	"github.com/jfperren/Peerster/common"
	"math/rand"
	"sync"
	"time"
)

type Mixer struct {
	ToSend chan *common.OnionPacket
	buffer [common.MixerNodeBufferSize]*common.OnionPacket // buffer to contain packets, and send them when the buffer is filled
	bufferSize uint

	lock sync.RWMutex
}

func NewMixer() *Mixer {
	var m Mixer
	m.bufferSize = 0
	m.ToSend = make(chan *common.OnionPacket)
	return &m
}

func (m *Mixer) ForwardPacket(p *common.OnionPacket) {
	m.lock.Lock()
	m.buffer[m.bufferSize] = p
	m.bufferSize++
	if m.bufferSize == common.MixerNodeBufferSize {
		m.lock.Unlock()
		m.ReleasePackets()
	} else {
		m.lock.Unlock()
	}
}

func (m *Mixer) ReleasePackets() {
	m.lock.Lock()
	var randomDuration = time.Duration(rand.Intn(common.MixerRandomTimeSleepRange)) * time.Millisecond
	time.Sleep(randomDuration)
	for _, packet := range m.buffer {
		m.ToSend <- packet
	}
	m.bufferSize = 0
	m.lock.Unlock()
}