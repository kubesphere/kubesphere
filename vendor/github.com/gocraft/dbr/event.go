package dbr

// EventReceiver gets events from dbr methods for logging purposes
type EventReceiver interface {
	Event(eventName string)
	EventKv(eventName string, kvs map[string]string)
	EventErr(eventName string, err error) error
	EventErrKv(eventName string, err error, kvs map[string]string) error
	Timing(eventName string, nanoseconds int64)
	TimingKv(eventName string, nanoseconds int64, kvs map[string]string)
}

type kvs map[string]string

var nullReceiver = &NullEventReceiver{}

// NullEventReceiver is a sentinel EventReceiver; use it if the caller doesn't supply one
type NullEventReceiver struct{}

// Event receives a simple notification when various events occur
func (n *NullEventReceiver) Event(eventName string) {}

// EventKv receives a notification when various events occur along with
// optional key/value data
func (n *NullEventReceiver) EventKv(eventName string, kvs map[string]string) {}

// EventErr receives a notification of an error if one occurs
func (n *NullEventReceiver) EventErr(eventName string, err error) error { return err }

// EventErrKv receives a notification of an error if one occurs along with
// optional key/value data
func (n *NullEventReceiver) EventErrKv(eventName string, err error, kvs map[string]string) error {
	return err
}

// Timing receives the time an event took to happen
func (n *NullEventReceiver) Timing(eventName string, nanoseconds int64) {}

// TimingKv receives the time an event took to happen along with optional key/value data
func (n *NullEventReceiver) TimingKv(eventName string, nanoseconds int64, kvs map[string]string) {}
