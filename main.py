#!/usr/bin/env python3
import sys
from commands.root import cli
from commands.diff import diff
from commands.sync import sync
from commands.audit import audit
from commands.snapshot import snapshot
from commands.rollback import rollback
from commands.validate import validate

cli.add_command(diff)
cli.add_command(sync)
cli.add_command(audit)
cli.add_command(snapshot)
cli.add_command(rollback)
cli.add_command(validate)

if __name__ == "__main__":
    cli()
    
