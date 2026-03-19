package telemetry

import (
	"context"
	"log/slog"
)

// multiHandler sends log records to multiple handlers
type multiHandler struct {
	handlers []slog.Handler
}

func (m multiHandler) Enabled(ctx context.Context, level slog.Level) bool {
	for _, h := range m.handlers {
		if h.Enabled(ctx, level) {
			return true
		}
	}
	return false
}

func (m multiHandler) Handle(ctx context.Context, record slog.Record) error {
	for _, h := range m.handlers {
		if err := h.Handle(ctx, record); err != nil {
			return err
		}
	}
	return nil
}

func (m multiHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithAttrs(attrs)
	}
	return multiHandler{handlers}
}

func (m multiHandler) WithGroup(name string) slog.Handler {
	handlers := make([]slog.Handler, len(m.handlers))
	for i, h := range m.handlers {
		handlers[i] = h.WithGroup(name)
	}
	return multiHandler{handlers}
}

// levelFilterHandler wraps a handler and filters logs below a certain level
type levelFilterHandler struct {
	handler slog.Handler
	level   slog.Level
}

func (h *levelFilterHandler) Enabled(ctx context.Context, level slog.Level) bool {
	return level >= h.level
}

func (h *levelFilterHandler) Handle(ctx context.Context, record slog.Record) error {
	return h.handler.Handle(ctx, record)
}

func (h *levelFilterHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	return &levelFilterHandler{
		handler: h.handler.WithAttrs(attrs),
		level:   h.level,
	}
}

func (h *levelFilterHandler) WithGroup(name string) slog.Handler {
	return &levelFilterHandler{
		handler: h.handler.WithGroup(name),
		level:   h.level,
	}
}
