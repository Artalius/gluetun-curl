package openvpn

import (
	"context"

	"github.com/qdm12/gluetun/internal/configuration/settings"
)

type Runner struct {
	settings settings.OpenVPN
	starter  CmdStarter
	logger   Logger
}

func NewRunner(settings settings.OpenVPN, starter CmdStarter,
	logger Logger,
) *Runner {
	return &Runner{
		starter:  starter,
		logger:   logger,
		settings: settings,
	}
}

func (r *Runner) Run(ctx context.Context, errCh chan<- error, ready chan<- struct{}) {
	stdoutLines, stderrLines, waitError, err := start(ctx, r.starter, r.settings.Version, r.settings.Flags)
	if err != nil {
		errCh <- err
		return
	}

	streamCtx, streamCancel := context.WithCancel(context.Background())
	streamDone := make(chan struct{})
	go streamLines(streamCtx, streamDone, r.logger,
		stdoutLines, stderrLines, ready)

	select {
	case <-ctx.Done():
		<-waitError
		streamCancel()
		<-streamDone
		errCh <- ctx.Err()
	case err := <-waitError:
		streamCancel()
		<-streamDone
		errCh <- err
	}
}
