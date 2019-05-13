package quic

import (
	"fmt"
	"sync"

	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/wire"
)

//go:generate genny -in $GOFILE -out streams_map_incoming_bidi.go gen "item=streamI Item=BidiStream streamTypeGeneric=protocol.StreamTypeBidi"
//go:generate genny -in $GOFILE -out streams_map_incoming_uni.go gen "item=receiveStreamI Item=UniStream streamTypeGeneric=protocol.StreamTypeUni"
type incomingItemsMap struct {
	mutex sync.RWMutex
	cond  sync.Cond

	streams map[protocol.StreamID]item
	// When a stream is deleted before it was accepted, we can't delete it immediately.
	// We need to wait until the application accepts it, and delete it immediately then.
	streamsToDelete map[protocol.StreamID]struct{} // used as a set

	nextStreamToAccept protocol.StreamID // the next stream that will be returned by AcceptStream()
	nextStreamToOpen   protocol.StreamID // the highest stream that the peer openend
	maxStream          protocol.StreamID // the highest stream that the peer is allowed to open
	maxNumStreams      uint64            // maximum number of streams

	newStream        func(protocol.StreamID) item
	queueMaxStreamID func(*wire.MaxStreamsFrame)

	closeErr error
}

func newIncomingItemsMap(
	nextStreamToAccept protocol.StreamID,
	initialMaxStreamID protocol.StreamID,
	maxNumStreams uint64,
	queueControlFrame func(wire.Frame),
	newStream func(protocol.StreamID) item,
) *incomingItemsMap {
	m := &incomingItemsMap{
		streams:            make(map[protocol.StreamID]item),
		streamsToDelete:    make(map[protocol.StreamID]struct{}),
		nextStreamToAccept: nextStreamToAccept,
		nextStreamToOpen:   nextStreamToAccept,
		maxStream:          initialMaxStreamID,
		maxNumStreams:      maxNumStreams,
		newStream:          newStream,
		queueMaxStreamID:   func(f *wire.MaxStreamsFrame) { queueControlFrame(f) },
	}
	m.cond.L = &m.mutex
	return m
}

func (m *incomingItemsMap) AcceptStream() (item, error) {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	var id protocol.StreamID
	var str item
	for {
		id = m.nextStreamToAccept
		var ok bool
		if m.closeErr != nil {
			return nil, m.closeErr
		}
		str, ok = m.streams[id]
		if ok {
			break
		}
		m.cond.Wait()
	}
	m.nextStreamToAccept += 4
	// If this stream was completed before being accepted, we can delete it now.
	if _, ok := m.streamsToDelete[id]; ok {
		delete(m.streamsToDelete, id)
		if err := m.deleteStream(id); err != nil {
			return nil, err
		}
	}
	return str, nil
}

func (m *incomingItemsMap) GetOrOpenStream(id protocol.StreamID) (item, error) {
	m.mutex.RLock()
	if id > m.maxStream {
		m.mutex.RUnlock()
		return nil, fmt.Errorf("peer tried to open stream %d (current limit: %d)", id, m.maxStream)
	}
	// if the id is smaller than the highest we accepted
	// * this stream exists in the map, and we can return it, or
	// * this stream was already closed, then we can return the nil
	if id < m.nextStreamToOpen {
		var s item
		// If the stream was already queued for deletion, and is just waiting to be accepted, don't return it.
		if _, ok := m.streamsToDelete[id]; !ok {
			s = m.streams[id]
		}
		m.mutex.RUnlock()
		return s, nil
	}
	m.mutex.RUnlock()

	m.mutex.Lock()
	// no need to check the two error conditions from above again
	// * maxStream can only increase, so if the id was valid before, it definitely is valid now
	// * highestStream is only modified by this function
	for newID := m.nextStreamToOpen; newID <= id; newID += 4 {
		m.streams[newID] = m.newStream(newID)
		m.cond.Signal()
	}
	m.nextStreamToOpen = id + 4
	s := m.streams[id]
	m.mutex.Unlock()
	return s, nil
}

func (m *incomingItemsMap) DeleteStream(id protocol.StreamID) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	return m.deleteStream(id)
}

func (m *incomingItemsMap) deleteStream(id protocol.StreamID) error {
	if _, ok := m.streams[id]; !ok {
		return fmt.Errorf("Tried to delete unknown stream %d", id)
	}

	// Don't delete this stream yet, if it was not yet accepted.
	// Just save it to streamsToDelete map, to make sure it is deleted as soon as it gets accepted.
	if id >= m.nextStreamToAccept {
		if _, ok := m.streamsToDelete[id]; ok {
			return fmt.Errorf("Tried to delete stream %d multiple times", id)
		}
		m.streamsToDelete[id] = struct{}{}
		return nil
	}

	delete(m.streams, id)
	// queue a MAX_STREAM_ID frame, giving the peer the option to open a new stream
	if m.maxNumStreams > uint64(len(m.streams)) {
		numNewStreams := m.maxNumStreams - uint64(len(m.streams))
		m.maxStream = m.nextStreamToOpen + protocol.StreamID((numNewStreams-1)*4)
		m.queueMaxStreamID(&wire.MaxStreamsFrame{
			Type:       streamTypeGeneric,
			MaxStreams: m.maxStream.StreamNum(),
		})
	}
	return nil
}

func (m *incomingItemsMap) CloseWithError(err error) {
	m.mutex.Lock()
	m.closeErr = err
	for _, str := range m.streams {
		str.closeForShutdown(err)
	}
	m.mutex.Unlock()
	m.cond.Broadcast()
}
