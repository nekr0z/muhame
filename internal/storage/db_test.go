package storage_test

import (
	"context"
	"embed"
	"errors"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/stretchr/testify/assert"
	"github.com/testcontainers/testcontainers-go"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"github.com/testcontainers/testcontainers-go/wait"
	"go.uber.org/zap"

	"github.com/nekr0z/muhame/internal/metrics"
	"github.com/nekr0z/muhame/internal/storage"
)

//go:embed testdata/migrations/*.sql
var fs embed.FS

var testStorage storage.Storage

const (
	testCounterName = "test_counter"
	testGaugeName   = "test_gauge"
)

var (
	testCounterValue = metrics.Counter(11)
	testGaugeValue   = metrics.Gauge(1.2)
)

func TestGet(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		want    metrics.Metric
		wantErr bool
	}{
		{
			name: testCounterName,
			want: testCounterValue,
		},
		{
			name: testGaugeName,
			want: testGaugeValue,
		},
		{
			name:    "no_counter",
			want:    metrics.Counter(0),
			wantErr: true,
		},
		{
			name:    "no_gauge",
			want:    metrics.Gauge(0),
			wantErr: true,
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			got, err := testStorage.Get(ctx, tt.want.Type(), tt.name)

			if tt.wantErr {
				assert.True(t, errors.Is(err, storage.ErrMetricNotFound))
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.want, got)
			}
		})
	}
}

func TestUpdate(t *testing.T) {
	var (
		newCounterName = "another_counter"
		newGaugeName   = "another_gauge"
	)
	tests := []struct {
		name      string
		prevValue metrics.Metric
		mName     string
		mValue    metrics.Metric
	}{
		{
			name:   "new counter",
			mName:  newCounterName,
			mValue: metrics.Counter(5),
		},
		{
			name:   "new gauge",
			mName:  newGaugeName,
			mValue: metrics.Gauge(25.18),
		},
		{
			name:   "existing counter",
			mName:  newCounterName,
			mValue: metrics.Counter(999),
		},
		{
			name:   "existing gauge",
			mName:  newGaugeName,
			mValue: metrics.Gauge(18.4),
		},
	}

	ctx := context.Background()

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			want := tt.mValue

			have, err := testStorage.Get(ctx, tt.mValue.Type(), tt.mName)
			if !errors.Is(err, storage.ErrMetricNotFound) {
				want, err = have.Update(want)
				assert.NoError(t, err)
			}

			err = testStorage.Update(ctx, metrics.Named{
				Name:   tt.mName,
				Metric: tt.mValue,
			})
			assert.NoError(t, err)

			got, err := testStorage.Get(ctx, tt.mValue.Type(), tt.mName)
			assert.NoError(t, err)

			assert.Equal(t, want, got)
		})
	}
}

func TestList(t *testing.T) {
	t.Parallel()

	ctx := context.Background()

	ms, err := testStorage.List(ctx)
	assert.NoError(t, err)

	assert.Contains(t, ms, metrics.Named{Name: testCounterName, Metric: testCounterValue})
	assert.Contains(t, ms, metrics.Named{Name: testGaugeName, Metric: testGaugeValue})
}

func TestMain(m *testing.M) {
	ctx := context.Background()
	log := zap.NewNop()

	dbName := "metrics"
	dbUser := "user"
	dbPassword := "password"

	postgresContainer, err := postgres.Run(ctx,
		"postgres:17",
		postgres.WithDatabase(dbName),
		postgres.WithUsername(dbUser),
		postgres.WithPassword(dbPassword),
		testcontainers.WithWaitStrategy(
			wait.ForLog("database system is ready to accept connections").
				WithOccurrence(2).
				WithStartupTimeout(5*time.Second),
			wait.ForListeningPort("5432/tcp")),
	)
	if err != nil {
		panic(err)
	}

	defer func() {
		err = testcontainers.TerminateContainer(postgresContainer)
		if err != nil {
			panic(err)
		}
	}()

	dsn, err := postgresContainer.ConnectionString(ctx, "sslmode=disable")
	if err != nil {
		panic(err)
	}

	testStorage, err = storage.New(log.Sugar(), storage.Config{
		DatabaseDSN: dsn,
	})
	if err != nil {
		panic(err)
	}

	migrateTestData(dsn)

	m.Run()
}

func migrateTestData(dsn string) {
	src, err := iofs.New(fs, "testdata/migrations")
	if err != nil {
		panic(err)
	}

	m, err := migrate.NewWithSourceInstance("iofs", src, dsn)
	if err != nil {
		panic(err)
	}

	err = m.Up()
	if err != nil {
		panic(err)
	}
}
