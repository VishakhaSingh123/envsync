import sys
import click
from colorama import Fore
from core.parser import load_config
from core.snapshot import restore
from commands.root import cli, print_banner, success, info, warn, confirm_prompt


@cli.command()
@click.argument("env")
@click.option("--id", "snap_id", default=None, help="Snapshot ID to rollback to")
@click.pass_context
def rollback(ctx, env, snap_id):
    """Rollback environment to the last snapshot"""

    config_path = ctx.obj["config_path"]
    print_banner(f"Rollback: {env}")

    try:
        config = load_config(config_path)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to load config: {e}")
        sys.exit(1)

    warn(f"This will overwrite the current state of '{env}'.")
    if not confirm_prompt("Proceed with rollback?"):
        info("Rollback cancelled.")
        return

    try:
        snap = restore(config, env, snap_id)
        success(f"Rollback complete: '{env}' restored to snapshot {snap['id']}")
        info(f"Snapshot was taken at: {snap['created_at'][:19]}")
    except Exception as e:
        print(Fore.RED + f"✗ Rollback failed: {e}")
        sys.exit(1)


