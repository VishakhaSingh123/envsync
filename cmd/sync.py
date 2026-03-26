import sys
import click
from colorama import Fore
from core.parser import load_config, load_environment
from core.comparator import compare, has_drift
from core.comparator import print_table
from core.sync import build_plan, apply
from core.snapshot import create as create_snapshot
from commands.root import cli, print_banner, info, success, warn, confirm_prompt


@cli.command()
@click.argument("source")
@click.argument("target")
@click.option("--dry-run", is_flag=True, help="Preview changes without applying")
@click.option("--keys", default=None, help="Comma-separated keys to sync")
@click.option("--overwrite", is_flag=True, help="Overwrite conflicts without prompting")
@click.pass_context
def sync(ctx, source, target, dry_run, keys, overwrite):
    """Synchronize source environment into target"""

    config_path = ctx.obj["config_path"]
    strict = ctx.obj["strict"]
    print_banner(f"Sync: {source} → {target}")

    if strict and target in ["production", "prod"]:
        warn("STRICT MODE: Syncing to production requires approval.")
        if not confirm_prompt("Are you sure you want to sync to PRODUCTION?"):
            info("Sync cancelled.")
            return

    try:
        config = load_config(config_path)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to load config: {e}")
        sys.exit(1)

    try:
        src = load_environment(config, source)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to load source '{source}': {e}")
        sys.exit(1)

    try:
        tgt = load_environment(config, target)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to load target '{target}': {e}")
        sys.exit(1)

    entries = compare(src, tgt)

    if not has_drift(entries):
        success("Environments are already in sync. Nothing to do.")
        return

    print_table(entries, source, target)

    if dry_run:
        info("DRY RUN: No changes applied.")
        return

    info(f"Creating snapshot of '{target}' before sync...")
    try:
        snap = create_snapshot(config, target)
        success(f"Snapshot saved: {snap['id']}")
    except Exception as e:
        warn(f"Could not create snapshot: {e} (continuing anyway)")

    plan = build_plan(src, tgt, entries, keys_filter=keys, overwrite=overwrite)

    try:
        applied = apply(config, target, plan)
    except Exception as e:
        print(Fore.RED + f"✗ Sync failed: {e}")
        print(Fore.RED + f"Run: python main.py rollback {target} to restore previous state")
        sys.exit(1)

    print()
    success(f"Sync complete: {applied} keys applied to '{target}'")
