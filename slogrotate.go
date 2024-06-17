package slogrotate

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"sync"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// ensure we always implement slog.Handler
var _ slog.Handler = (*Handler)(nil)

type Handler struct {
	formatter slog.Handler
	logger    *lumberjack.Logger
	opts      *HandlerOptions

	once sync.Once
	mu   sync.Mutex

	lastTime time.Time
}

type HandlerOptions struct {
	// Filename is the file to write logs to.  Backup log files will be retained
	// in the same directory.  It uses <processname>-current.log in
	// os.TempDir() if empty.
	Filename string `json:"filename" yaml:"filename"`

	// MaxSize is the maximum size in megabytes of the log file before it gets
	// rotated. It defaults to 100 megabytes.
	MaxSize int `json:"maxsize" yaml:"maxsize"`

	// MaxAge is the maximum number of days to retain old log files based on the
	// timestamp encoded in their filename.  Note that a day is defined as 24
	// hours and may not exactly correspond to calendar days due to daylight
	// savings, leap seconds, etc. The default is not to remove old log files
	// based on age.
	MaxAge int `json:"maxage" yaml:"maxage"`

	// MaxBackups is the maximum number of old log files to retain.  The default
	// is to retain all old log files (though MaxAge may still cause them to get
	// deleted.)
	MaxBackups int `json:"maxbackups" yaml:"maxbackups"`

	// Frequency is the frequency of rotation.  The default is none
	Frequency Frequency `json:"frequency" yaml:"frequency"`

	// LocalTime determines if the time used for formatting the timestamps in
	// backup files is the computer's local time.  The default is to use UTC
	// time.
	LocalTime bool `json:"localtime" yaml:"localtime"`

	// Compress determines if the rotated log files should be compressed
	// using gzip. The default is not to perform compression.
	Compress bool `json:"compress" yaml:"compress"`

	HandlerOptions slog.HandlerOptions
	Builder        Builder
}

func (o *HandlerOptions) GetBuilder() Builder {
	b := o.Builder
	if b == nil {
		b = func(w io.Writer, opts *slog.HandlerOptions) slog.Handler {
			return slog.NewJSONHandler(w, opts)
		}
	}
	return b
}

// GetFilename generates the name of the logfile from the current time.
func (o *HandlerOptions) GetFilename() string {
	if o.Filename != "" {
		return o.Filename
	}
	name := filepath.Base(os.Args[0]) + "-current.log"
	return filepath.Join(os.TempDir(), name)
}

// HandlerBuilder is a type representing functions used to create
// handlers to control formatting of logging data.
type HandlerBuilder[H slog.Handler] func(w io.Writer, opts *slog.HandlerOptions) H
type Builder func(w io.Writer, opts *slog.HandlerOptions) slog.Handler

// NewHandler creates a new handler with the given options.
func NewHandler(opts *HandlerOptions) *Handler {
	if opts == nil {
		opts = &HandlerOptions{}
	}

	logger := &lumberjack.Logger{
		Filename:   opts.GetFilename(),
		MaxSize:    opts.MaxSize,
		MaxAge:     opts.MaxAge,
		MaxBackups: opts.MaxBackups,
		LocalTime:  opts.LocalTime,
		Compress:   opts.Compress,
	}

	return &Handler{
		formatter: opts.GetBuilder()(logger, &opts.HandlerOptions),
		logger:    logger,
		opts:      opts,
	}
}

// Close implements io.Closer, and closes the current logfile.
func (h *Handler) Close() error {
	return h.logger.Close()
}

func (h *Handler) clone() *Handler {
	return &Handler{
		formatter: h.formatter,
		logger:    h.logger,
		opts:      h.opts,
		lastTime:  h.lastTime,
	}
}

func (h *Handler) Enabled(ctx context.Context, level slog.Level) bool {
	return h.formatter.Enabled(ctx, level)
}

// Handle implements the method of the slog.Handler interface.  If the current
// date is different from the previous date, the file is closed, renamed to
// include a timestamp of the current time, and a new log file is created
// using the original log file name. If failed, an error is returned.
func (h *Handler) Handle(ctx context.Context, record slog.Record) error {
	if h.opts.Frequency != FrequencyNone {
		h.once.Do(func() { h.initTime(record.Time) })

		if !h.opts.Frequency.IsSame(h.lastTime, record.Time) {
			h.mu.Lock()
			defer h.mu.Unlock()

			if !h.opts.Frequency.IsSame(h.lastTime, record.Time) {
				h.lastTime = record.Time
				return h.logger.Rotate()
			}
		}
	}

	return h.formatter.Handle(ctx, record)
}

func (h *Handler) initTime(t time.Time) {
	filename := h.opts.GetFilename()
	stat, err := os.Stat(filename)
	if err == nil {
		h.lastTime = stat.ModTime()
	} else {
		h.lastTime = t
	}
}

// WithAttrs implements the method of the slog.Handler interface by
// cloning the current handler and calling the WithAttrs of the
// formatter handler.
func (h *Handler) WithAttrs(attr []slog.Attr) slog.Handler {
	nh := h.clone()
	nh.formatter = h.formatter.WithAttrs(attr)
	return nh
}

// WithGroup implements the method of the slog.Handler interface by
// cloning the current handler and calling the WithGroup of the
// formatter handler.
func (h *Handler) WithGroup(name string) slog.Handler {
	nh := h.clone()
	nh.formatter = h.formatter.WithGroup(name)
	return nh
}
