package storage

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"time"

	"go.uber.org/zap"

	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/internal/retry"
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

// Update implements the Storage interface.
func (fs *fileStorage) Update(ctx context.Context, m metrics.Named) error {
	if err := fs.s.Update(ctx, m); err != nil {
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

// List returns all metrics.
func (fs *fileStorage) List(ctx context.Context) ([]metrics.Named, error) {
	return fs.s.List(ctx)
}

// Get returns a metric by name.
func (fs *fileStorage) Get(ctx context.Context, t, name string) (metrics.Metric, error) {
	return fs.s.Get(ctx, t, name)
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
	f, err := retry.OnError(func() (*os.File, error) {
		return os.Open(fs.c.Filename)
	}, func(err error) bool {
		return err != nil
	})
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer f.Close()

	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		jm := scanner.Bytes()
		named, err := metrics.FromJSON(jm)
		if err != nil {
			return fmt.Errorf("failed to parse json: %w", err)
		}

		err = fs.s.Update(ctx, named)
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
	f, err := retry.OnError(func() (*os.File, error) {
		return os.Create(fs.c.Filename)
	}, func(err error) bool {
		return err != nil
	})
	if err != nil {
		return fmt.Errorf("failed to create/truncate file: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	nameds, err := fs.s.List(ctx)
	if err != nil {
		return fmt.Errorf("failed to list metrics: %w", err)
	}

	for _, named := range nameds {
		jm := metrics.ToJSON(named.Metric, named.Name)

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
