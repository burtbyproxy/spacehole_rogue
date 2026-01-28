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
// Long messages are wrapped at maxWidth (default 55 chars for comms panel).
func (l *MessageLog) Add(text string, priority MsgPriority) {
	const maxWidth = 55
	lines := wrapText(text, maxWidth)
	for _, line := range lines {
		msg := Message{Text: line, Priority: priority}
		if len(l.Messages) >= l.maxSize {
			copy(l.Messages, l.Messages[1:])
			l.Messages[len(l.Messages)-1] = msg
		} else {
			l.Messages = append(l.Messages, msg)
		}
	}
}

// wrapText splits text into lines no longer than maxWidth.
func wrapText(s string, maxWidth int) []string {
	if len(s) <= maxWidth {
		return []string{s}
	}
	var result []string
	words := splitWords(s)
	if len(words) == 0 {
		return []string{""}
	}
	line := words[0]
	for _, w := range words[1:] {
		if len(line)+1+len(w) > maxWidth {
			result = append(result, line)
			line = w
		} else {
			line += " " + w
		}
	}
	if line != "" {
		result = append(result, line)
	}
	return result
}

// splitWords splits on whitespace.
func splitWords(s string) []string {
	var words []string
	word := ""
	for _, r := range s {
		if r == ' ' || r == '\t' || r == '\n' {
			if word != "" {
				words = append(words, word)
				word = ""
			}
		} else {
			word += string(r)
		}
	}
	if word != "" {
		words = append(words, word)
	}
	return words
}

// Recent returns the last n messages (or fewer if the log is shorter).
func (l *MessageLog) Recent(n int) []Message {
	if n > len(l.Messages) {
		n = len(l.Messages)
	}
	return l.Messages[len(l.Messages)-n:]
}
