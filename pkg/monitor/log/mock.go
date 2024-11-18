package log

import "context"

type LogMock struct {
	Log

	onRead func()
	err    error
}

func NewLogMock(err error) *LogMock {
	return &LogMock{
		err: err,
	}
}

func (m *LogMock) Stream(ctx context.Context, lh LogHandler) error {
	if m.onRead != nil {
		m.onRead()
	}
	return m.err
}

func (m *LogMock) SetErr(newErr error) {
	m.err = newErr
}

func (m *LogMock) SetOnRead(onRead func()) {
	m.onRead = onRead
}
