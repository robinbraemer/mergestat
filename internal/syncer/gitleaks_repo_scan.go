package syncer

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"

	"github.com/jackc/pgx/v4"
	"github.com/mergestat/mergestat/internal/db"
	"github.com/mergestat/mergestat/internal/helper"
)

// handleGitleaksRepoScan executes `gitleaks detect {git-repo} -f json` for a repo
// and inserts the output JSON into the DB
func (w *worker) handleGitleaksRepoScan(ctx context.Context, j *db.DequeueSyncJobRow) error {
	l := w.loggerForJob(j)

	tmpPath, cleanup, err := helper.CreateTempDir(os.Getenv("GIT_CLONE_PATH"), "mergestat-repo-")
	if err != nil {
		return fmt.Errorf("temp dir: %w", err)
	}
	defer func() {
		if err := cleanup(); err != nil {
			l.Err(err).Msgf("error cleaning up repo at: %s, %v", tmpPath, err)
		}
	}()

	var ghToken string
	if ghToken, err = w.fetchGitHubTokenFromDB(ctx); err != nil {
		return err
	}

	if err = w.cloneRepo(ctx, ghToken, j.Repo, tmpPath, false, j); err != nil {
		return fmt.Errorf("git clone: %w", err)
	}

	// indicate that we're starting a gitleaks scan
	if err := w.sendBatchLogMessages(ctx, []*syncLog{{Type: SyncLogTypeInfo, RepoSyncQueueID: j.ID,
		Message: fmt.Sprintf(LogFormatStartingSync, j.SyncType, j.Repo),
	}}); err != nil {
		return fmt.Errorf("send batch log messages: %w", err)
	}

	cmd := exec.CommandContext(ctx, "gitleaks", "detect", "-s", tmpPath, "-f", "json", "-r", "_mergestat_gitleaks_scan_results.json", "--exit-code", "0")

	if err = cmd.Run(); err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			w.logger.Warn().AnErr("error", exitErr).Str("stderr", string(exitErr.Stderr)).Msgf("error running gitleaks scan")
		}
		return fmt.Errorf("running gitleaks scan: %w", err)
	}

	var output []byte
	if output, err = os.ReadFile("_mergestat_gitleaks_scan_results.json"); err != nil {
		return fmt.Errorf("reading gitleaks scan results: %w", err)
	}

	var tx pgx.Tx
	if tx, err = w.pool.BeginTx(ctx, pgx.TxOptions{}); err != nil {
		return fmt.Errorf("begin tx: %w", err)
	}
	defer func() {
		if err := tx.Rollback(ctx); err != nil {
			if !errors.Is(err, pgx.ErrTxClosed) {
				w.logger.Err(err).Msgf("could not rollback transaction")
			}
		}
	}()

	r, err := tx.Exec(ctx, "DELETE FROM gitleaks_repo_scans WHERE repo_id = $1;", j.RepoID.String())
	if err != nil {
		return fmt.Errorf("exec delete: %w", err)
	}

	if err := w.sendBatchLogMessages(ctx, []*syncLog{{
		Type:            SyncLogTypeInfo,
		RepoSyncQueueID: j.ID,
		Message:         fmt.Sprintf("removed %d row(s) from gitleaks_repo_scans", r.RowsAffected()),
	}}); err != nil {
		return err
	}

	if _, err := tx.Exec(ctx, "INSERT INTO gitleaks_repo_scans (repo_id, results) VALUES ($1, $2)", j.RepoID, output); err != nil {
		return fmt.Errorf("inserting gitleaks results: %w", err)
	}

	l.Info().Msg("inserted gitleaks scan results")

	if err := w.sendBatchLogMessages(ctx, []*syncLog{{
		Type:            SyncLogTypeInfo,
		RepoSyncQueueID: j.ID,
		Message:         "inserted gitleaks scan results into gitleaks_repo_scans",
	}}); err != nil {
		return err
	}

	if err := w.db.WithTx(tx).SetSyncJobStatus(ctx, db.SetSyncJobStatusParams{Status: "DONE", ID: j.ID}); err != nil {
		return fmt.Errorf("update status done: %w", err)
	}

	// indicate that we're finishing query execution
	if err := w.sendBatchLogMessages(ctx, []*syncLog{{Type: SyncLogTypeInfo, RepoSyncQueueID: j.ID,
		Message: fmt.Sprintf(LogFormatFinishingSync, j.SyncType, j.Repo),
	}}); err != nil {
		return fmt.Errorf("send batch log messages: %w", err)
	}

	return tx.Commit(ctx)
}
