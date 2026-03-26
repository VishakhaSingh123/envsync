import sys
import click
from colorama import Fore
from core.parser import load_config
from core.snapshot import create
from commands.root import cli, print_banner, success, info

@cli.command()
@click.argument("env")
@click.pass_context
def snapshot(ctx, env):
    """Take a snapshot of an environment"""

    config_path = ctx.obj["config_path"]
    print_banner(f"Snapshot: {env}")

    try:
        config = load_config(config_path)
    except Exception as e:
        print(Fore.RED + f"[ERROR] Failed to load config: {e}")
        sys.exit(1)

    try:
        snap = create(config, env)
        success(f"Snapshot created: {snap['id']}")
        info(f"Saved at: {snap['created_at'][:19]}")
    except Exception as e:
        print(Fore.RED + f"[ERROR] Snapshot failed: {e}")
        sys.exit(1)