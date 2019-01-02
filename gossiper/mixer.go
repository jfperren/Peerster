package gossiper

import (
	"github.com/jfperren/Peerster/common"
	"sync"
)

type Mixer struct {
	buffer [common.MixerNodeBufferSize]*common.OnionPacket // buffer to contain packets, and send them when the buffer is filled
	bufferSize uint

	lock sync.RWMutex
}

func NewMixer() *Mixer {
	var m Mixer
	m.bufferSize = 0
	return &m
}

func (m *Mixer) ForwardPacket(p *common.OnionPacket) {
	m.lock.Lock()
	m.buffer[m.bufferSize] = p
	m.bufferSize++
	if m.bufferSize == common.MixerNodeBufferSize {
		defer m.ReleasePackets()
	}
	m.lock.Unlock()
}

func (m *Mixer) ReleasePackets() {
	for _, packet := range m.buffer {
		// TODO: send onion to next address in mix network
		_ = packet
	}
	m.bufferSize = 0
}