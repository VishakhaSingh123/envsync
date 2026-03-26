import sys
import click
from colorama import Fore
from core.parser import load_config, load_environment, load_source_of_truth
from commands.root import cli, print_banner


@cli.command()
@click.option("--env", default="dev", help="Environment to audit")
@click.option("--fail-on-missing", is_flag=True, help="Exit 1 if any keys missing")
@click.option("--threshold", default=0, help="Max allowed drift count")
@click.pass_context
def audit(ctx, env, fail_on_missing, threshold):
    """Audit a single environment for missing or undefined keys"""

    config_path = ctx.obj["config_path"]
    print_banner(f"Audit: {env}")

    try:
        config = load_config(config_path)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to load config: {e}")
        sys.exit(1)

    try:
        environment = load_environment(config, env)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to load environment '{env}': {e}")
        sys.exit(1)

    try:
        truth = load_source_of_truth(config)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to load source of truth: {e}")
        sys.exit(1)

    missing = []
    empty = []
    extra = []
    ok = 0

    for key in truth:
        if key not in environment:
            missing.append(key)
        elif not environment[key] or environment[key] in ["CHANGE_ME", "TODO"]:
            empty.append(key)
        else:
            ok += 1

    for key in environment:
        if key not in truth:
            extra.append(key)

    if missing:
        print(Fore.RED + f"\n✗ MISSING keys (not in {env}):")
        for k in missing:
            print(Fore.RED + f"   - {k}")

    if empty:
        print(Fore.YELLOW + "\n⚠ EMPTY / PLACEHOLDER keys:")
        for k in empty:
            print(Fore.YELLOW + f"   ~ {k}")

    if extra:
        print(Fore.BLUE + "\nℹ EXTRA keys (not in .env.example):")
        for k in extra:
            print(Fore.BLUE + f"   + {k}")

    print()
    print(Fore.GREEN + f"✔ OK: {ok} keys")
    print(Fore.RED + f"✗ Missing: {len(missing)} keys")
    print(Fore.YELLOW + f"~ Empty: {len(empty)} keys")
    print(Fore.BLUE + f"+ Extra: {len(extra)} keys")

    total_drift = len(missing) + len(empty)

    if threshold > 0 and total_drift > threshold:
        print(Fore.RED + f"\n✗ AUDIT FAILED: Drift count {total_drift} exceeds threshold {threshold}")
        sys.exit(1)

    if fail_on_missing and missing:
        print(Fore.RED + "\n✗ AUDIT FAILED: Missing keys detected")
        sys.exit(1)

    if total_drift == 0:
        print(Fore.GREEN + f"\n✔ Environment '{env}' is clean ✓")
    else:
        print(Fore.YELLOW + f"\n⚠ Environment '{env}' has {total_drift} issues")
