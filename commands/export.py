import sys
import json
from datetime import datetime

import click
from colorama import Fore

from core.parser import load_config, load_environment
from core.snapshot import list_snapshots
from commands.root import cli, print_banner, success, warn


@cli.command()
@click.option("--output", "-o", default="web/dashboard/data.json",
              help="Path to write the dashboard JSON data file")
@click.pass_context
def export(ctx, output):
    """Export live environments + snapshots as JSON for the dashboard"""

    config_path = ctx.obj["config_path"]
    print_banner("Export Dashboard Data")

    try:
        config = load_config(config_path)
    except Exception as e:
        print(Fore.RED + f"[ERROR] Failed to load config: {e}")
        sys.exit(1)

    env_names = list(config.get("environments", {}).keys())
    if not env_names:
        print(Fore.RED + "[ERROR] No environments defined in envsync.yaml")
        sys.exit(1)

    environments = {}
    for env_name in env_names:
        try:
            environments[env_name] = {"keys": load_environment(config, env_name)}
        except Exception as e:
            warn(f"Skipping '{env_name}': {e}")

    snapshots = []
    for env_name in env_names:
        for snap in list_snapshots(config, env_name):
            snapshots.append({
                "id": snap.get("id", ""),
                "env": snap.get("env", env_name),
                "ts": snap.get("created_at", ""),
                "keys": snap.get("key_count", 0),
                "note": "",
            })
    snapshots.sort(key=lambda s: s["ts"], reverse=True)

    data = {
        "generated_at": datetime.now().isoformat(),
        "environments": environments,
        "snapshots": snapshots,
        "logs": [],
    }

    try:
        with open(output, "w") as f:
            json.dump(data, f, indent=2)
    except Exception as e:
        print(Fore.RED + f"[ERROR] Failed to write {output}: {e}")
        sys.exit(1)

    success(f"Exported {len(environments)} environment(s), {len(snapshots)} snapshot(s) to {output}")
    print(Fore.CYAN + "[INFO] Open web/dashboard/index.html (served, not file://) to see it live.")