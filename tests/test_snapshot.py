import os
import pytest
from core.snapshot import create, list_snapshots, restore, prune_old, snapshot_dir


@pytest.fixture
def config(tmp_path):
    env_file = tmp_path / ".env.dev"
    env_file.write_text("DB_HOST=localhost\nDB_PORT=5432\n")
    snap_dir = tmp_path / ".envsync" / "snapshots"
    return {
        "environments": {
            "dev": {"type": "file", "path": str(env_file), "format": "dotenv"}
        },
        "snapshots": {"directory": str(snap_dir), "max_keep": 10, "encrypted": False},
        "secrets": {"encryption_key_env": "ENVSYNC_TEST_KEY"},
    }


class TestCreate:
    def test_creates_snapshot_file(self, config):
        snap = create(config, "dev")
        assert os.path.exists(snap["path"])

    def test_snapshot_contains_env_data(self, config):
        snap = create(config, "dev")
        assert snap["data"]["DB_HOST"] == "localhost"
        assert snap["key_count"] == 2

    def test_snapshot_id_has_expected_prefix(self, config):
        snap = create(config, "dev")
        assert snap["id"].startswith("snap_")


class TestListSnapshots:
    def test_empty_when_no_snapshots(self, config):
        assert list_snapshots(config, "dev") == []

    def test_lists_created_snapshot(self, config):
        create(config, "dev")
        snaps = list_snapshots(config, "dev")
        assert len(snaps) == 1

    def test_sorted_newest_first(self, config):
        create(config, "dev")
        create(config, "dev")
        snaps = list_snapshots(config, "dev")
        # regression check: rapid creates must not collide on the same snapshot id
        assert len(snaps) == 2
        assert snaps[0]["created_at"] >= snaps[1]["created_at"]


class TestRestore:
    def test_restore_writes_back_to_env_file(self, config):
        create(config, "dev")
        env_path = config["environments"]["dev"]["path"]
        # mutate the file to simulate drift
        with open(env_path, "w") as f:
            f.write("DB_HOST=changed\n")

        restore(config, "dev")
        content = open(env_path).read()
        assert "localhost" in content

    def test_restore_no_snapshots_raises(self, config):
        with pytest.raises(ValueError):
            restore(config, "dev")

    def test_restore_specific_snapshot_id(self, config):
        snap = create(config, "dev")
        result = restore(config, "dev", snap_id=snap["id"])
        assert result["id"] == snap["id"]

    def test_restore_unknown_id_raises(self, config):
        create(config, "dev")
        with pytest.raises(ValueError):
            restore(config, "dev", snap_id="snap_doesnotexist")


class TestPruneOld:
    def test_prunes_beyond_max_keep(self, config):
        config["snapshots"]["max_keep"] = 2
        for _ in range(4):
            create(config, "dev")
        snaps = list_snapshots(config, "dev")
        assert len(snaps) <= 2

    def test_max_keep_zero_disables_pruning(self, config):
        config["snapshots"]["max_keep"] = 0
        for _ in range(3):
            create(config, "dev")
        snaps = list_snapshots(config, "dev")
        assert len(snaps) == 3