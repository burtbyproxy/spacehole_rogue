package game

// MsgPriority controls the color of a message in the comms log.
type MsgPriority uint8

const (
	MsgInfo     MsgPriority = iota // cyan
	MsgWarning                     // yellow
	MsgCritical                    // red
	MsgDiscovery                   // green
	MsgSocial                      // white
)

// Message is a single entry in the comms log.
type Message struct {
	Text     string
	Priority MsgPriority
}

// MessageLog is a bounded FIFO of messages.
type MessageLog struct {
	Messages []Message
	maxSize  int
}

// NewMessageLog creates a log that keeps the most recent maxSize messages.
func NewMessageLog(maxSize int) *MessageLog {
	return &MessageLog{
		Messages: make([]Message, 0, maxSize),
		maxSize:  maxSize,
	}
}

// Add appends a message, evicting the oldest if full.
func (l *MessageLog) Add(text string, priority MsgPriority) {
	msg := Message{Text: text, Priority: priority}
	if len(l.Messages) >= l.maxSize {
		copy(l.Messages, l.Messages[1:])
		l.Messages[len(l.Messages)-1] = msg
	} else {
		l.Messages = append(l.Messages, msg)
	}
}

// Recent returns the last n messages (or fewer if the log is shorter).
func (l *MessageLog) Recent(n int) []Message {
	if n > len(l.Messages) {
		n = len(l.Messages)
	}
	return l.Messages[len(l.Messages)-n:]
}
