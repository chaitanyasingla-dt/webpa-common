package device

import (
	"github.com/Comcast/webpa-common/wrp"
)

// EventType is the type of device-related event
type EventType uint8

const (
	Connect EventType = iota
	Disconnect
	MessageReceived
	MessageFailed
	Pong

	InvalidEventString string = "!!INVALID DEVICE EVENT TYPE!!"
)

var (
	// emptyString is a convenient instance of an empty string
	emptyString string
)

func (et EventType) String() string {
	switch et {
	case Connect:
		return "Connect"
	case Disconnect:
		return "Disconnect"
	case MessageReceived:
		return "MessageReceived"
	case MessageFailed:
		return "MessageFailed"
	case Pong:
		return "Pong"
	default:
		return InvalidEventString
	}
}

// Event represents a single occurrence of interest for device-related applications.
// Instances of Event should be considered immutable by application code.  Also, Event
// instances should not be stored across calls to a listener, as the infrastructure is
// free to reuse Event instances.
type Event struct {
	// Type describes the kind of this event.  This field is always set.
	Type EventType

	// Device refers to the device, possibly disconnected, for which this event is being set.
	// This field is always set.
	Device Interface

	// Message is the WRP message relevant to this event.  This field is only set for
	// MessageReceived and MessageFailed events.
	//
	// Never assume that it is safe to use this Message outside the listener invocation.  Make
	// a copy if this Message is needed by other goroutines or if it needs to be part of a long-lived
	// data structure.
	Message wrp.Routable

	// Encoded is the encoded representation of the Message field.  It is always set if and only if
	// the Message field is set.
	//
	// Never assume that it is safe to use this byte slice outside the listener invocation.  Make
	// a copy if this byte slice is needed by other goroutines or if it needs to be part of a long-lived
	// data structure.
	Encoded []byte

	// Data is the pong data associated with this event.  This field is only set for a Pong event.
	Data string
}

// setMessageFailed sets or resets this event's fields to represent a MessageFailed event.
func (e *Event) setMessageFailed(device Interface, message wrp.Routable, encoded []byte) {
	e.Type = MessageFailed
	e.Device = device
	e.Message = message
	e.Encoded = encoded
	e.Data = emptyString
}

// setMessageReceived sets or resets this event's fields to represent a MessageReceived event.
func (e *Event) setMessageReceived(device Interface, message wrp.Routable, encoded []byte) {
	e.Type = MessageReceived
	e.Device = device
	e.Message = message
	e.Encoded = encoded
	e.Data = emptyString
}

// setPong sets or resets this event's fields to represent a Pong event.
func (e *Event) setPong(device Interface, data string) {
	e.Type = Pong
	e.Device = device
	e.Message = nil
	e.Encoded = nil
	e.Data = data
}

// Listener is an event sink.  Listeners should never modify events and should never
// store events for later use.  If data from an event is needed for another goroutine
// or for long-term storage, a copy should be made.
type Listener func(*Event)
