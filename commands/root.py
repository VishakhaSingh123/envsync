import click
from colorama import Fore, Style, init

init(autoreset=True)

BANNER = """
+-------------------------------------------------------+
|              EnvSync v1.0 - Config Parity Tool         |
|   Detect drift - Sync secrets - Validate runtimes      |
+-------------------------------------------------------+
"""


@click.group()
@click.option("--config", "-c", default="envsync.yaml", help="Path to envsync config file")
@click.option("--verbose", "-v", is_flag=True, help="Verbose output")
@click.option("--strict", is_flag=True, help="Strict mode for production")
@click.pass_context
def cli(ctx, config, verbose, strict):
    """EnvSync - Environment Synchronization Tool"""
    ctx.ensure_object(dict)
    ctx.obj["config_path"] = config
    ctx.obj["verbose"] = verbose
    ctx.obj["strict"] = strict


def success(msg):
    print(Fore.GREEN + "[OK] " + msg)


def error(msg):
    print(Fore.RED + "[ERROR] " + msg)


def info(msg):
    print(Fore.CYAN + "[INFO] " + msg)


def warn(msg):
    print(Fore.YELLOW + "[WARN] " + msg)


def print_banner(title):
    print()
    print(Fore.CYAN + "== " + title + " ==")
    print()


def confirm_prompt(question):
    answer = input(f"{question} [y/N]: ").strip().lower()
    return answer == "y"
