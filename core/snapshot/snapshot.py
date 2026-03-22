import os
import json
from datetime import datetime
from core.parser import load_environment, write_environment
from core.crypto import get_encryption_key, encrypt_map, decrypt_map


def snapshot_dir(config, env_name):
    base = config.get("snapshots", {}).get("directory", ".envsync/snapshots")
    return os.path.join(base, env_name)


def create(config, env_name):
    env = load_environment(config, env_name)
    snap_id = f"snap_{datetime.now().strftime('%Y%m%d_%H%M%S')}"
    directory = snapshot_dir(config, env_name)
    os.makedirs(directory, mode=0o700, exist_ok=True)

    snap = {
        "id": snap_id,
        "env": env_name,
        "created_at": datetime.now().isoformat(),
        "key_count": len(env),
        "data": env
    }

    encrypted = config.get("snapshots", {}).get("encrypted", False)
    if encrypted:
        try:
            key_env = config.get("secrets", {}).get("encryption_key_env", "ENVSYNC_KEY")
            key = get_encryption_key(key_env)
            snap["data"] = encrypt_map(env, key)
        except Exception:
            pass

    path = os.path.join(directory, f"{snap_id}.json")
    with open(path, "w") as f:
        json.dump(snap, f, indent=2)

    prune_old(config, env_name)
    snap["path"] = path
    return snap


def list_snapshots(config, env_name):
    directory = snapshot_dir(config, env_name)
    if not os.path.exists(directory):
        return []

    snaps = []
    for fname in os.listdir(directory):
        if not fname.endswith(".json"):
            continue
        path = os.path.join(directory, fname)
        try:
            with open(path) as f:
                s = json.load(f)
            s["path"] = path
            snaps.append(s)
        except Exception:
            continue

    snaps.sort(key=lambda x: x.get("created_at", ""), reverse=True)
    return snaps


def restore(config, env_name, snap_id=None):
    snaps = list_snapshots(config, env_name)
    if not snaps:
        raise ValueError(f"No snapshots found for '{env_name}'")

    target = None
    if not snap_id:
        target = snaps[0]
    else:
        for s in snaps:
            if s["id"] == snap_id:
                target = s
                break

    if not target:
        raise ValueError(f"Snapshot '{snap_id}' not found")

    data = target["data"]

    encrypted = config.get("snapshots", {}).get("encrypted", False)
    if encrypted:
        try:
            key_env = config.get("secrets", {}).get("encryption_key_env", "ENVSYNC_KEY")
            key = get_encryption_key(key_env)
            data = decrypt_map(data, key)
        except Exception:
            pass

    write_environment(config, env_name, data)
    return target


def prune_old(config, env_name):
    max_keep = config.get("snapshots", {}).get("max_keep", 10)
    if max_keep <= 0:
        return
    snaps = list_snapshots(config, env_name)
    for old in snaps[max_keep:]:
        try:
            os.remove(old["path"])
        except Exception:
            pass
