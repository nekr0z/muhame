package storage

import (
	"bufio"
	"fmt"
	"os"
	"time"

	"github.com/nekr0z/muhame/internal/metrics"
	"go.uber.org/zap"
)

type Config struct {
	Interval time.Duration
	Filename string
	Restore  bool
}

var _ PersistentStorage = &FileStorage{}

type FileStorage struct {
	c                  Config
	s                  Storage
	stopChan, doneChan chan struct{}
}

func NewFileStorage(log *zap.SugaredLogger, c Config) *FileStorage {
	fs := &FileStorage{
		c:        c,
		s:        NewMemStorage(),
		stopChan: make(chan struct{}),
		doneChan: make(chan struct{}),
	}

	if c.Restore {
		fs.load(log)
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
				fs.save(log)
				break loop
			case <-time.After(interval):
				if c.Interval == 0 {
					continue
				}
				fs.save(log)
			}
		}
		close(fs.doneChan)
	}()

	return fs
}

func (fs *FileStorage) Update(name string, m metrics.Metric) error {
	if err := fs.s.Update(name, m); err != nil {
		return err
	}

	if fs.c.Interval == 0 {
		err := fs.flush()
		if err != nil {
			return fmt.Errorf("failed to save metrics to file: %w", err)
		}
	}

	return nil
}

func (fs *FileStorage) List() ([]string, []metrics.Metric, error) {
	return fs.s.List()
}

func (fs *FileStorage) Get(t, name string) (metrics.Metric, error) {
	return fs.s.Get(t, name)
}

// Flush breaks the flushing loop and blocks until metrics are saved to file (or
// failed to do that).
func (fs *FileStorage) Flush() {
	close(fs.stopChan)
	<-fs.doneChan
}

func (fs *FileStorage) load(log *zap.SugaredLogger) {
	log.Infof("restoring from file %s", fs.c.Filename)
	if err := fs.restore(); err != nil {
		log.Errorf("failed to restore metrics from file: %s", err)
	}
}

func (fs *FileStorage) restore() error {
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

		err = fs.s.Update(name, m)
		if err != nil {
			return fmt.Errorf("failed to update metric: %w", err)
		}
	}

	return nil
}

func (fs *FileStorage) save(log *zap.SugaredLogger) {
	if err := fs.flush(); err != nil {
		log.Errorf("failed to save metrics to file: %s", err)
	}
	log.Infof("metrics saved to file")
}

func (fs *FileStorage) flush() error {
	f, err := os.Create(fs.c.Filename)
	if err != nil {
		return fmt.Errorf("failed to create/truncate file: %w", err)
	}
	defer f.Close()

	w := bufio.NewWriter(f)
	defer w.Flush()

	names, ms, err := fs.s.List()
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
