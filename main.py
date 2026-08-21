import sys
from commands.root import cli
from commands.diff import diff
from commands.sync import sync
from commands.audit import audit
from commands.snapshot import snapshot
from commands.rollback import rollback
from commands.validate import validate
from commands.export import export

cli.add_command(diff)
cli.add_command(sync)
cli.add_command(audit)
cli.add_command(snapshot)
cli.add_command(rollback)
cli.add_command(validate)
cli.add_command(export)

if __name__ == "__main__":
    cli()