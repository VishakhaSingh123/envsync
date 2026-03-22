import json
import yaml
from tabulate import tabulate
from colorama import Fore, Style, init

init(autoreset=True)

SENSITIVE_WORDS = ["password", "secret", "token", "key", "private", "credential", "auth"]


def is_sensitive(key):
    lower = key.lower()
    return any(word in lower for word in SENSITIVE_WORDS)


def mask_value(value):
    if len(value) <= 4:
        return "****"
    return value[:2] + "*" * (len(value) - 4) + value[-2:]


def display_value(key, value):
    if not value:
        return "(empty)"
    if is_sensitive(key):
        return mask_value(value)
    return value


def compare(source, target):
    entries = []
    all_keys = set(list(source.keys()) + list(target.keys()))

    for key in all_keys:
        in_src = key in source
        in_tgt = key in target

        if in_src and not in_tgt:
            entries.append({
                "key": key,
                "status": "MISSING",
                "source_value": source[key],
                "target_value": ""
            })
        elif not in_src and in_tgt:
            entries.append({
                "key": key,
                "status": "EXTRA",
                "source_value": "",
                "target_value": target[key]
            })
        elif source[key] != target[key]:
            entries.append({
                "key": key,
                "status": "MISMATCH",
                "source_value": source[key],
                "target_value": target[key]
            })
        else:
            entries.append({
                "key": key,
                "status": "MATCH",
                "source_value": source[key],
                "target_value": target[key]
            })

    priority = {"MISSING": 0, "MISMATCH": 1, "EXTRA": 2, "MATCH": 3}
    entries.sort(key=lambda x: (priority[x["status"]], x["key"]))
    return entries


def has_drift(entries):
    return any(e["status"] != "MATCH" for e in entries)


def missing_count(entries):
    return sum(1 for e in entries if e["status"] == "MISSING")


def mismatch_count(entries):
    return sum(1 for e in entries if e["status"] == "MISMATCH")


def extra_count(entries):
    return sum(1 for e in entries if e["status"] == "EXTRA")


def print_table(entries, src_name, tgt_name):
    rows = []
    for e in entries:
        if e["status"] == "MATCH":
            continue
        key = e["key"]
        status = e["status"]
        src_val = display_value(key, e["source_value"])
        tgt_val = display_value(key, e["target_value"])

        if status == "MISSING":
            status_str = Fore.RED + "MISSING" + Style.RESET_ALL
        elif status == "MISMATCH":
            status_str = Fore.YELLOW + "MISMATCH" + Style.RESET_ALL
        elif status == "EXTRA":
            status_str = Fore.BLUE + "EXTRA" + Style.RESET_ALL
        else:
            status_str = Fore.GREEN + "MATCH" + Style.RESET_ALL

        rows.append([key, status_str, src_val, tgt_val])

    headers = ["KEY", "STATUS", src_name, tgt_name]
    print(tabulate(rows, headers=headers, tablefmt="simple"))

    matches = sum(1 for e in entries if e["status"] == "MATCH")
    if matches > 0:
        print(Fore.GREEN + f"\n  + {matches} keys match (hidden)")


def print_json(entries):
    print(json.dumps({"entries": entries}, indent=2))


def print_yaml_output(entries):
    print(yaml.dump({"entries": entries}, default_flow_style=False))
