import sys
import subprocess
import click
from colorama import Fore
from core.parser import load_config, scaffold_config
from commands.root import cli, print_banner, success, info


def get_installed_version(runtime):
    commands = {
        "node": ["node", "--version"],
        "python": ["python3", "--version"],
        "python3": ["python3", "--version"],
        "go": ["go", "version"],
        "ruby": ["ruby", "--version"],
        "java": ["java", "-version"],
    }
    cmd = commands.get(runtime, [runtime, "--version"])
    try:
        result = subprocess.run(cmd, capture_output=True, text=True)
        output = (result.stdout or result.stderr).strip()
        for part in output.split():
            part = part.lstrip("v")
            if part and part[0].isdigit():
                return part
        return output
    except FileNotFoundError:
        return None


def version_matches(installed, required):
    installed = installed.lstrip("v")
    required = required.lstrip("v")
    return installed.startswith(required)


@cli.command()
@click.option("--env", default="dev", help="Environment to validate")
@click.pass_context
def validate(ctx, env):
    """Validate runtime versions against required spec"""

    config_path = ctx.obj["config_path"]
    print_banner(f"Runtime Validation: {env}")

    try:
        config = load_config(config_path)
    except Exception as e:
        print(Fore.RED + f"✗ Failed to load config: {e}")
        sys.exit(1)

    runtimes = config.get("runtimes", {})
    if not runtimes:
        info("No runtimes defined in envsync.yaml. Skipping.")
        return

    all_pass = True
    for runtime, required in runtimes.items():
        installed = get_installed_version(runtime)
        if not installed:
            print(Fore.RED + f"  ✗ {runtime:<12} required: {required:<12} installed: not found")
            all_pass = False
        elif version_matches(installed, required):
            print(Fore.GREEN + f"  ✔ {runtime:<12} required: {required:<12} installed: {installed}")
        else:
            print(Fore.RED + f"  ✗ {runtime:<12} required: {required:<12} installed: {installed}")
            all_pass = False

    print()
    if all_pass:
        success("All runtime versions match!")
    else:
        print(Fore.RED + "✗ Runtime version mismatch detected.")


@cli.command("init")
@click.pass_context
def init_cmd(ctx):
    """Scaffold a new envsync.yaml config file"""

    print_banner("Init EnvSync Project")
    try:
        scaffold_config("envsync.yaml")
        success("Created envsync.yaml")
        info("Edit this file then run: python main.py audit --env dev")
    except Exception as e:
        print(Fore.RED + f"✗ Failed to scaffold config: {e}")
        sys.exit(1)
