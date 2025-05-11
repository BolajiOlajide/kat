package output

import (
	tea "github.com/charmbracelet/bubbletea"
)

const (
	padding  = 2
	maxWidth = 80
)

func NewProgressWriter(p *tea.Program) *ProgressWriter {
	return &ProgressWriter{
		p: p,
		onProgress: func(ratio float64) {
			p.Send(ProgresCount(ratio))
		},
	}
}

type ProgresCount float64

func (p ProgresCount) IsComplete() bool {
	return p >= 100
}

type ProgressError struct {
	Err error
}

type ProgressWriter struct {
	total      int
	current    int
	p          *tea.Program
	onProgress func(float64)
}

func (pw *ProgressWriter) Start() {
	// No-op since we're not dealing with files anymore
}

func (pw *ProgressWriter) Increment(amount int) {
	pw.current += amount
	if pw.total > 0 && pw.onProgress != nil {
		pw.onProgress(float64(pw.current) / float64(pw.total))
	}
}

func (pw *ProgressWriter) SetTotal(total int) {
	pw.total = total
}

func (pw *ProgressWriter) SetCurrent(current int) {
	pw.current = current
	if pw.total > 0 && pw.onProgress != nil {
		pw.onProgress(float64(pw.current) / float64(pw.total))
	}
}

func (pw *ProgressWriter) Reset() {
	pw.current = 0
	if pw.total > 0 && pw.onProgress != nil {
		pw.onProgress(0)
	}
}

func (pw *ProgressWriter) Write(p []byte) (int, error) {
	pw.current += len(p)
	if pw.total > 0 && pw.onProgress != nil {
		pw.onProgress(float64(pw.current) / float64(pw.total))
	}
	return len(p), nil
}
