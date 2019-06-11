package flowcontrol

import (
	"fmt"

	"github.com/lucas-clemente/quic-go/internal/congestion"
	"github.com/lucas-clemente/quic-go/internal/protocol"
	"github.com/lucas-clemente/quic-go/internal/qerr"
	"github.com/lucas-clemente/quic-go/internal/utils"
)

type streamFlowController struct {
	baseFlowController

	streamID protocol.StreamID

	queueWindowUpdate func()

	connection connectionFlowControllerI

	receivedFinalOffset bool
}

var _ StreamFlowController = &streamFlowController{}

// NewStreamFlowController gets a new flow controller for a stream
func NewStreamFlowController(
	streamID protocol.StreamID,
	cfc ConnectionFlowController,
	receiveWindow protocol.ByteCount,
	maxReceiveWindow protocol.ByteCount,
	initialSendWindow protocol.ByteCount,
	queueWindowUpdate func(protocol.StreamID),
	rttStats *congestion.RTTStats,
	logger utils.Logger,
) StreamFlowController {
	return &streamFlowController{
		streamID:          streamID,
		connection:        cfc.(connectionFlowControllerI),
		queueWindowUpdate: func() { queueWindowUpdate(streamID) },
		baseFlowController: baseFlowController{
			rttStats:             rttStats,
			receiveWindow:        receiveWindow,
			receiveWindowSize:    receiveWindow,
			maxReceiveWindowSize: maxReceiveWindow,
			sendWindow:           initialSendWindow,
			logger:               logger,
		},
	}
}

// UpdateHighestReceived updates the highestReceived value, if the offset is higher.
func (c *streamFlowController) UpdateHighestReceived(offset protocol.ByteCount, final bool) error {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// If the final offset for this stream is already known, check for consistency.
	if c.receivedFinalOffset {
		// If we receive another final offset, check that it's the same.
		if final && offset != c.highestReceived {
			return qerr.Error(qerr.FinalSizeError, fmt.Sprintf("Received inconsistent final offset for stream %d (old: %#x, new: %#x bytes)", c.streamID, c.highestReceived, offset))
		}
		// Check that the offset is below the final offset.
		if offset > c.highestReceived {
			return qerr.Error(qerr.FinalSizeError, fmt.Sprintf("Received offset %#x for stream %d. Final offset was already received at %#x", offset, c.streamID, c.highestReceived))
		}
	}

	if final {
		c.receivedFinalOffset = true
	}
	if offset == c.highestReceived {
		return nil
	}
	// A higher offset was received before.
	// This can happen due to reordering.
	if offset <= c.highestReceived {
		if final {
			return qerr.Error(qerr.FinalSizeError, fmt.Sprintf("Received final offset %#x for stream %d, but already received offset %#x before", offset, c.streamID, c.highestReceived))
		}
		return nil
	}

	increment := offset - c.highestReceived
	c.highestReceived = offset
	if c.checkFlowControlViolation() {
		return qerr.Error(qerr.FlowControlError, fmt.Sprintf("Received %#x bytes on stream %d, allowed %#x bytes", offset, c.streamID, c.receiveWindow))
	}
	return c.connection.IncrementHighestReceived(increment)
}

func (c *streamFlowController) AddBytesRead(n protocol.ByteCount) {
	c.baseFlowController.AddBytesRead(n)
	c.maybeQueueWindowUpdate()
	c.connection.AddBytesRead(n)
}

func (c *streamFlowController) Abandon() {
	if unread := c.highestReceived - c.bytesRead; unread > 0 {
		c.connection.AddBytesRead(unread)
	}
}

func (c *streamFlowController) AddBytesSent(n protocol.ByteCount) {
	c.baseFlowController.AddBytesSent(n)
	c.connection.AddBytesSent(n)
}

func (c *streamFlowController) SendWindowSize() protocol.ByteCount {
	return utils.MinByteCount(c.baseFlowController.sendWindowSize(), c.connection.SendWindowSize())
}

func (c *streamFlowController) maybeQueueWindowUpdate() {
	c.mutex.Lock()
	hasWindowUpdate := !c.receivedFinalOffset && c.hasWindowUpdate()
	c.mutex.Unlock()
	if hasWindowUpdate {
		c.queueWindowUpdate()
	}
}

func (c *streamFlowController) GetWindowUpdate() protocol.ByteCount {
	// don't use defer for unlocking the mutex here, GetWindowUpdate() is called frequently and defer shows up in the profiler
	c.mutex.Lock()
	// if we already received the final offset for this stream, the peer won't need any additional flow control credit
	if c.receivedFinalOffset {
		c.mutex.Unlock()
		return 0
	}

	oldWindowSize := c.receiveWindowSize
	offset := c.baseFlowController.getWindowUpdate()
	if c.receiveWindowSize > oldWindowSize { // auto-tuning enlarged the window size
		c.logger.Debugf("Increasing receive flow control window for stream %d to %d kB", c.streamID, c.receiveWindowSize/(1<<10))
		c.connection.EnsureMinimumWindowSize(protocol.ByteCount(float64(c.receiveWindowSize) * protocol.ConnectionFlowControlMultiplier))
	}
	c.mutex.Unlock()
	return offset
}
