import sys
import click
from core.parser import load_config, load_environment
from core.comparator import compare, has_drift, print_table, print_json, print_yaml_output
from core.comparator import missing_count, mismatch_count, extra_count
from commands.root import cli, print_banner
from colorama import Fore


@cli.command()
@click.argument("env1")
@click.argument("env2")
@click.option("--output", "-o", default="table", type=click.Choice(["table", "json", "yaml"]),
              help="Output format")
@click.pass_context
def diff(ctx, env1, env2, output):
    """Compare two environments and show drift report"""

    config_path = ctx.obj["config_path"]
    print_banner(f"Drift Report: {env1} → {env2}")

    try:
        config = load_config(config_path)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to load config: {e}")
        sys.exit(1)

    try:
        source = load_environment(config, env1)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to load environment '{env1}': {e}")
        sys.exit(1)

    try:
        target = load_environment(config, env2)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to load environment '{env2}': {e}")
        sys.exit(1)

    entries = compare(source, target)

    if output == "json":
        print_json(entries)
    elif output == "yaml":
        print_yaml_output(entries)
    else:
        print_table(entries, env1, env2)

    if has_drift(entries):
        print(Fore.YELLOW + f"\n⚠  Drift detected: {missing_count(entries)} missing, "
              f"{mismatch_count(entries)} mismatched, {extra_count(entries)} extra")
        sys.exit(2)
    else:
        print(Fore.GREEN + "\n✔  Environments are in sync!")

