package main

import (
	"context"
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"strings"
	"unicode"
	"unicode/utf8"
)

type IRCLogHandler struct {
	level slog.Level
}

func NewIRCLogHandler(level slog.Level) *IRCLogHandler {
	return &IRCLogHandler{
		level: level,
	}
}

func (h *IRCLogHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *IRCLogHandler) Handle(_ context.Context, r slog.Record) error {
	var sb strings.Builder

	sb.WriteString("PRIVMSG #logs :")

	sb.WriteString(r.Message)

	r.Attrs(func(a slog.Attr) bool {
		sb.WriteString(" ")

		key := a.Key
		if needsQuoting(key) {
			key = strconv.Quote(key)
		}
		sb.WriteString(key)
		sb.WriteString("=")

		val := attrValueToString(a.Value)
		if needsQuoting(val) {
			val = strconv.Quote(val)
		}
		sb.WriteString(val)

		return true
	})

	str := sb.String()

	select {
	case ircSendBuffered <- str:
	default:
		fmt.Fprintln(os.Stderr, "DROP")
	}

	fmt.Fprintln(os.Stderr, str)

	return nil
}

func (h *IRCLogHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return h
}

func (h *IRCLogHandler) WithGroup(_ string) slog.Handler {
	return h
}

func attrValueToString(v slog.Value) string {
	return v.String()
}

func init() {
	slog.SetDefault(slog.New(NewIRCLogHandler(slog.LevelInfo)))
}

// copied from slog
func needsQuoting(s string) bool {
	if len(s) == 0 {
		return true
	}
	for i := 0; i < len(s); {
		b := s[i]
		if b < utf8.RuneSelf {
			if b != '\\' && (b == ' ' || b == '=' || !safeSet[b]) {
				return true
			}
			i++
			continue
		}
		r, size := utf8.DecodeRuneInString(s[i:])
		if r == utf8.RuneError || unicode.IsSpace(r) || !unicode.IsPrint(r) {
			return true
		}
		i += size
	}
	return false
}

var safeSet = [256]bool{
	'!': true, '#': true, '$': true, '%': true, '&': true, '\'': true,
	'*': true, '+': true, ',': true, '-': true, '.': true, '/': true,
	'0': true, '1': true, '2': true, '3': true, '4': true,
	'5': true, '6': true, '7': true, '8': true, '9': true,
	':': true, ';': true, '<': true, '>': true, '?': true, '@': true,
	'A': true, 'B': true, 'C': true, 'D': true, 'E': true, 'F': true,
	'G': true, 'H': true, 'I': true, 'J': true, 'K': true, 'L': true,
	'M': true, 'N': true, 'O': true, 'P': true, 'Q': true, 'R': true,
	'S': true, 'T': true, 'U': true, 'V': true, 'W': true, 'X': true,
	'Y': true, 'Z': true, '[': true, ']': true, '^': true, '_': true,
	'a': true, 'b': true, 'c': true, 'd': true, 'e': true, 'f': true,
	'g': true, 'h': true, 'i': true, 'j': true, 'k': true, 'l': true,
	'm': true, 'n': true, 'o': true, 'p': true, 'q': true, 'r': true,
	's': true, 't': true, 'u': true, 'v': true, 'w': true, 'x': true,
	'y': true, 'z': true, '{': true, '|': true, '}': true, '~': true,
}
