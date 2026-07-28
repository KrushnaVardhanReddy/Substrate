package workers

import (
	"context"
	"io/ioutil"
	"os"

	"github.com/KrushnaVardhanReddy/substrate/api/internal/db"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/fuzzer"
	"github.com/riverqueue/river"
	"github.com/KrushnaVardhanReddy/substrate/api/internal/db/sqlcgen"
)

type FuzzerJob struct {
	SpecContent string `json:"spec_content"`
}

func (FuzzerJob) Kind() string { return "fuzzer_job" }

type FuzzerWorker struct {
	river.WorkerDefaults[FuzzerJob]
	Store db.Store
}

func (w *FuzzerWorker) Work(ctx context.Context, job *river.Job[FuzzerJob]) error {
	tmpFile, err := ioutil.TempFile("", "openapi-*.yaml")
	if err != nil {
		return err
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write([]byte(job.Args.SpecContent)); err != nil {
		return err
	}
	tmpFile.Close()

	q := sqlcgen.New(w.Store.Pool())
	f := fuzzer.NewFuzzer(tmpFile.Name(), q)

	results := make(chan fuzzer.VulnerabilityReport, 10)

	f.RunBackgroundFuzzing(ctx, results)

	// Consume the channel to prevent goroutine leak
	for range results {
		// Vulnerabilities are already reported/inserted in RunBackgroundFuzzing inside fuzzer
	}

	return nil
}
