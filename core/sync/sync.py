from colorama import Fore, Style, init
from core.parser import load_environment, write_environment
from core.comparator import compare, has_drift

init(autoreset=True)

SENSITIVE_WORDS = ["password", "secret", "token", "key", "private", "credential"]


def mask_if_sensitive(key, value):
    lower = key.lower()
    if any(w in lower for w in SENSITIVE_WORDS):
        if len(value) <= 4:
            return "****"
        return value[:2] + "*" * (len(value) - 4) + value[-2:]
    return value


def resolve_conflict(key, src_val, tgt_val):
    print()
    print(Fore.YELLOW + f"CONFLICT: {key}")
    print(Fore.CYAN + f"  [S] Source value: {mask_if_sensitive(key, src_val)}")
    print(Fore.BLUE + f"  [T] Target value: {mask_if_sensitive(key, tgt_val)}")
    print(Fore.WHITE + "  [K] Keep target (skip)")

    choice = input("Choose [S/T/K]: ").strip().upper()

    if choice == "S":
        return src_val
    elif choice == "T":
        return tgt_val
    else:
        return None


def build_plan(src, tgt, entries, keys_filter=None, overwrite=False):
    plan = {"changes": {}}
    filter_keys = set()

    if keys_filter:
        filter_keys = {k.strip() for k in keys_filter.split(",") if k.strip()}

    for entry in entries:
        key = entry["key"]
        status = entry["status"]

        if filter_keys and key not in filter_keys:
            continue

        if status == "MISSING":
            plan["changes"][key] = entry["source_value"]

        elif status == "MISMATCH":
            if overwrite:
                plan["changes"][key] = entry["source_value"]
            else:
                val = resolve_conflict(key, entry["source_value"], entry["target_value"])
                if val is not None:
                    plan["changes"][key] = val

    return plan


def apply(config, target_name, plan):
    current = load_environment(config, target_name)

    for k, v in plan["changes"].items():
        current[k] = v

    write_environment(config, target_name, current)
    return len(plan["changes"])
