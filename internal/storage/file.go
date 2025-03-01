package storage

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"github.com/nekr0z/muhame/internal/metrics"
	"go.uber.org/zap"
)

type Config struct {
	Interval    time.Duration
	Filename    string
	Restore     bool
	DatabaseDSN string
}

type fileStorage struct {
	c                  Config
	s                  Storage
	stopChan, doneChan chan struct{}
}

func newFileStorage(ctx context.Context, log *zap.SugaredLogger, c Config) *fileStorage {
	fs := &fileStorage{
		c:        c,
		s:        newMemStorage(),
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}

	if c.Restore {
		fs.load(ctx, log)
	}

	go func() {
		interval := c.Interval
		if interval == 0 {
			interval = 24 * time.Hour
		}

	loop:
		for {
			select {
			case <-fs.stopChan:
				fs.save(ctx, log)
				break loop
			case <-time.After(interval):
				if c.Interval == 0 {
					continue
				}
				fs.save(ctx, log)
			}
		}
		close(fs.doneChan)
	}()

	return fs
}

func (fs *fileStorage) Update(ctx context.Context, name string, m metrics.Metric) error {
	if err := fs.s.Update(ctx, name, m); err != nil {
		return err
	}

	if fs.c.Interval == 0 {
		err := fs.flush(ctx)
		if err != nil {
			return fmt.Errorf("failed to save metrics to file: %w", err)
		}
	}

	return nil
}

func (fs *fileStorage) List(ctx context.Context) ([]string, []metrics.Metric, error) {
	return fs.s.List(ctx)
}

func (fs *fileStorage) Get(ctx context.Context, t, name string) (metrics.Metric, error) {
	return fs.s.Get(ctx, t, name)
}

func (fs *fileStorage) Ping(ctx context.Context) error {
	return fs.s.Ping(ctx)
}

// Close breaks the flushing loop and blocks until metrics are saved to file (or
// failed to do that).
func (fs *fileStorage) Close() {
	close(fs.stopChan)
	<-fs.doneChan
	fs.s.Close()
}

func (fs *fileStorage) load(ctx context.Context, log *zap.SugaredLogger) {
	log.Infof("restoring from file %s", fs.c.Filename)
	if err := fs.restore(ctx); err != nil {
		log.Errorf("failed to restore metrics from file: %s", err)
	}
}

func (fs *fileStorage) restore(ctx context.Context) error {
	f, err := os.Open(fs.c.Filename)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		jm := scanner.Bytes()
		name, m, err := metrics.FromJSON(jm)
		if err != nil {
			return fmt.Errorf("failed to parse json: %w", err)
		}

		err = fs.s.Update(ctx, name, m)
		if err != nil {
			return fmt.Errorf("failed to update metric: %w", err)
		}
	}

	return nil
}

func (fs *fileStorage) save(ctx context.Context, log *zap.SugaredLogger) {
	if err := fs.flush(ctx); err != nil {
		log.Errorf("failed to save metrics to file: %s", err)
	}
	log.Infof("metrics saved to file")
}

func (fs *fileStorage) flush(ctx context.Context) error {
	f, err := os.Create(fs.c.Filename)
	if err != nil {
		return fmt.Errorf("failed to create/truncate file: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	names, ms, err := fs.s.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list metrics: %w", err)
	}

	for i, name := range names {
		jm := metrics.ToJSON(ms[i], name)

		if err := writeLine(w, jm); err != nil {
			return fmt.Errorf("failed to write to file: %w", err)
		}
	}

	return nil
}

func writeLine(w *bufio.Writer, b []byte) error {
	_, err := w.Write(b)
	if err != nil {
		return err
	}

	_, err = w.WriteString("\n")
	return err
}
