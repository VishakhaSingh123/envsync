package snapshot

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"time"

	"github.com/envsync/internal/crypto"
	"github.com/envsync/internal/parser"
)

type Snapshot struct {
	ID        string            `json:"id"`
	Env       string            `json:"env"`
	CreatedAt time.Time         `json:"created_at"`
	KeyCount  int               `json:"key_count"`
	Data      map[string]string `json:"data"`
	Path      string            `json:"-"`
}

func snapshotDir(cfg *parser.Config, envName string) string {
	return filepath.Join(cfg.Snapshots.Directory, envName)
}

func Create(cfg *parser.Config, envName string) (*Snapshot, error) {
	env, err := parser.LoadEnvironment(cfg, envName)
	if err != nil {
		return nil, err
	}

	snap := &Snapshot{
		ID:        fmt.Sprintf("snap_%s", time.Now().Format("20060102_150405")),
		Env:       envName,
		CreatedAt: time.Now(),
		KeyCount:  len(env),
		Data:      env,
	}

	dir := snapshotDir(cfg, envName)
	if err := os.MkdirAll(dir, 0700); err != nil {
		return nil, fmt.Errorf("could not create snapshot dir: %w", err)
	}

	snap.Path = filepath.Join(dir, snap.ID+".json")

	var dataToWrite interface{} = snap

	if cfg.Snapshots.Encrypted {
		key, err := crypto.GetEncryptionKey(cfg.Secrets.EncryptionKey)
		if err == nil {
			encData, err := crypto.EncryptMap(env, key)
			if err == nil {
				snapCopy := *snap
				snapCopy.Data = encData
				dataToWrite = snapCopy
			}
		}
	}

	f, err := os.OpenFile(snap.Path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	enc := json.NewEncoder(f)
	enc.SetIndent("", "  ")
	if err := enc.Encode(dataToWrite); err != nil {
		return nil, err
	}

	// Prune old snapshots
	pruneOld(cfg, envName)

	return snap, nil
}

func List(cfg *parser.Config, envName string) ([]*Snapshot, error) {
	dir := snapshotDir(cfg, envName)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}

	var snaps []*Snapshot
	for _, e := range entries {
		if filepath.Ext(e.Name()) != ".json" {
			continue
		}
		s, err := loadSnapshot(filepath.Join(dir, e.Name()))
		if err != nil {
			continue
		}
		s.Path = filepath.Join(dir, e.Name())
		snaps = append(snaps, s)
	}

	sort.Slice(snaps, func(i, j int) bool {
		return snaps[i].CreatedAt.After(snaps[j].CreatedAt)
	})

	return snaps, nil
}

func Restore(cfg *parser.Config, envName, snapID string) (*Snapshot, error) {
	snaps, err := List(cfg, envName)
	if err != nil {
		return nil, err
	}
	if len(snaps) == 0 {
		return nil, fmt.Errorf("no snapshots found for environment '%s'", envName)
	}

	var target *Snapshot
	if snapID == "" {
		target = snaps[0] // Latest
	} else {
		for _, s := range snaps {
			if s.ID == snapID {
				target = s
				break
			}
		}
	}

	if target == nil {
		return nil, fmt.Errorf("snapshot '%s' not found", snapID)
	}

	data := target.Data

	// Decrypt if needed
	if cfg.Snapshots.Encrypted {
		key, err := crypto.GetEncryptionKey(cfg.Secrets.EncryptionKey)
		if err == nil {
			decData, err := crypto.DecryptMap(data, key)
			if err == nil {
				data = decData
			}
		}
	}

	if err := parser.WriteEnvironment(cfg, envName, data); err != nil {
		return nil, fmt.Errorf("failed to write restored state: %w", err)
	}

	return target, nil
}

func loadSnapshot(path string) (*Snapshot, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var s Snapshot
	if err := json.Unmarshal(data, &s); err != nil {
		return nil, err
	}
	return &s, nil
}

func pruneOld(cfg *parser.Config, envName string) {
	maxKeep := cfg.Snapshots.MaxKeep
	if maxKeep <= 0 {
		return
	}
	snaps, err := List(cfg, envName)
	if err != nil || len(snaps) <= maxKeep {
		return
	}
	for _, s := range snaps[maxKeep:] {
		os.Remove(s.Path)
	}
}
