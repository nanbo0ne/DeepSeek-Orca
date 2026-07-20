package openai

import "strings"

type thinkTag struct {
	open  string
	close string
}

var thinkTags = []thinkTag{
	{open: "<think>", close: "</think>"},
	{open: "<thinking>", close: "</thinking>"},
	{open: "<reasoning>", close: "</reasoning>"},
}

type thinkState int

const (
	thinkProbe thinkState = iota
	thinkInside
	thinkPassthrough
)

// thinkSplitter peels a leading tagged reasoning block out of the content stream
// into reasoning text. Some OpenAI-compatible providers inline thinking instead
// of populating reasoning_content. It only arms at the very start of the turn, so
// an answer that merely mentions these tags is never hijacked.
type thinkSplitter struct {
	state thinkState
	buf   string
	close string
}

func (t *thinkSplitter) push(s string) (reasoning, text string) {
	switch t.state {
	case thinkPassthrough:
		return "", s
	case thinkInside:
		return t.scanClose(s)
	}

	t.buf += s
	trimmed := strings.TrimLeft(t.buf, "\ufeff \t\r\n")
	lower := strings.ToLower(trimmed)
	if tag, ok := matchThinkTag(lower); ok {
		t.state = thinkInside
		t.close = tag.close
		t.buf = ""
		return t.scanClose(trimmed[len(tag.open):])
	}
	if couldBecomeThinkTag(lower) {
		return "", ""
	}
	return "", t.drainPassthrough()
}

func (t *thinkSplitter) scanClose(s string) (reasoning, text string) {
	t.buf += s
	closeTag := t.close
	if closeTag == "" {
		closeTag = "</think>"
	}
	if idx := indexFold(t.buf, closeTag); idx >= 0 {
		r := t.buf[:idx]
		rest := strings.TrimLeft(t.buf[idx+len(closeTag):], "\ufeff \t\r\n")
		t.buf = ""
		t.close = ""
		t.state = thinkPassthrough
		return r, rest
	}
	keep := markerSuffixLen(t.buf, closeTag)
	r := t.buf[:len(t.buf)-keep]
	t.buf = t.buf[len(t.buf)-keep:]
	return r, ""
}

// flush emits whatever is buffered when the stream ends mid-decision: an
// unterminated reasoning tag is reasoning; anything else is text.
func (t *thinkSplitter) flush() (reasoning, text string) {
	if t.buf == "" {
		return "", ""
	}
	out := t.buf
	t.buf = ""
	t.close = ""
	if t.state == thinkInside {
		return out, ""
	}
	return "", out
}

func (t *thinkSplitter) drainPassthrough() string {
	t.state = thinkPassthrough
	out := t.buf
	t.buf = ""
	return out
}

func matchThinkTag(lower string) (thinkTag, bool) {
	for _, tag := range thinkTags {
		if strings.HasPrefix(lower, tag.open) {
			return tag, true
		}
	}
	return thinkTag{}, false
}

func couldBecomeThinkTag(lower string) bool {
	if lower == "" {
		return true
	}
	for _, tag := range thinkTags {
		if len(lower) < len(tag.open) && strings.HasPrefix(tag.open, lower) {
			return true
		}
	}
	return false
}

func indexFold(s, marker string) int {
	return strings.Index(strings.ToLower(s), strings.ToLower(marker))
}

// markerSuffixLen returns the length of the longest proper suffix of s that is a
// prefix of marker - the tail to hold back in case the rest of the tag arrives in
// the next delta.
func markerSuffixLen(s, marker string) int {
	lowerS := strings.ToLower(s)
	lowerMarker := strings.ToLower(marker)
	max := len(marker) - 1
	if max > len(s) {
		max = len(s)
	}
	for n := max; n > 0; n-- {
		if strings.HasPrefix(lowerMarker, lowerS[len(lowerS)-n:]) {
			return n
		}
	}
	return 0
}
