
import sys
import click
from colorama import Fore
from core.parser import load_config
from core.snapshot import create, list_snapshots
from commands.root import cli, print_banner, success, info


@cli.group()
def snapshot():
    """Manage environment snapshots"""
    pass


@snapshot.command("create")
@click.argument("env")
@click.pass_context
def snapshot_create(ctx, env):
    """Create a snapshot of an environment"""

    config_path = ctx.obj["config_path"]
    print_banner(f"Snapshot: {env}")

    try:
        config = load_config(config_path)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to load config: {e}")
        sys.exit(1)

    try:
        snap = create(config, env)
        success(f"Snapshot created: {env} (ID: {snap['id']})")
        info(f"Saved to: {snap['path']}")
    except Exception as e:
        print(Fore.RED + f"✗ Failed to create snapshot: {e}")
        sys.exit(1)


@snapshot.command("list")
@click.argument("env")
@click.pass_context
def snapshot_list(ctx, env):
    """List all snapshots for an environment"""

    config_path = ctx.obj["config_path"]

    try:
        config = load_config(config_path)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to load config: {e}")
        sys.exit(1)

    print_banner(f"Snapshots for: {env}")

    try:
        snaps = list_snapshots(config, env)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to list snapshots: {e}")
        sys.exit(1)

    if not snaps:
        info("No snapshots found.")
        return

    for i, s in enumerate(snaps):
        marker = "▶ " if i == 0 else "  "
        print(Fore.GREEN + marker + f"[{s['id']}] {s['created_at'][:19]} — {s['key_count']} keys")
