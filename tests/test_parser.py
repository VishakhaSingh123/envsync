import os
import pytest
from core.parser import (
    parse_dotenv,
    parse_yaml_env,
    parse_json_env,
    load_config,
    load_from_file,
    write_dotenv,
)


class TestParseDotenv:
    def test_simple_key_value(self):
        content = "DB_HOST=localhost\nDB_PORT=5432"
        result = parse_dotenv(content)
        assert result == {"DB_HOST": "localhost", "DB_PORT": "5432"}

    def test_skips_comments_and_blank_lines(self):
        content = "# this is a comment\n\nDB_HOST=localhost\n  # indented comment\n"
        result = parse_dotenv(content)
        assert result == {"DB_HOST": "localhost"}

    def test_strips_surrounding_quotes(self):
        content = 'NAME="Vishakha"\nGREETING=\'hello world\''
        result = parse_dotenv(content)
        assert result == {"NAME": "Vishakha", "GREETING": "hello world"}

    def test_ignores_lines_without_equals(self):
        content = "DB_HOST=localhost\njust some garbage text"
        result = parse_dotenv(content)
        assert result == {"DB_HOST": "localhost"}

    def test_empty_value(self):
        content = "EMPTY_KEY="
        result = parse_dotenv(content)
        assert result == {"EMPTY_KEY": ""}

    def test_value_with_equals_sign(self):
        content = "CONNECTION_STRING=key=value;other=thing"
        result = parse_dotenv(content)
        assert result == {"CONNECTION_STRING": "key=value;other=thing"}


class TestParseYamlEnv:
    def test_basic_yaml(self):
        content = "DB_HOST: localhost\nDB_PORT: 5432"
        result = parse_yaml_env(content)
        assert result == {"DB_HOST": "localhost", "DB_PORT": "5432"}

    def test_values_coerced_to_string(self):
        content = "PORT: 5432\nDEBUG: true"
        result = parse_yaml_env(content)
        assert result["PORT"] == "5432"
        assert result["DEBUG"] == "True"

    def test_empty_yaml_returns_empty_dict(self):
        assert parse_yaml_env("") == {}


class TestParseJsonEnv:
    def test_basic_json(self):
        content = '{"DB_HOST": "localhost", "DB_PORT": 5432}'
        result = parse_json_env(content)
        assert result == {"DB_HOST": "localhost", "DB_PORT": "5432"}

    def test_invalid_json_raises(self):
        with pytest.raises(Exception):
            parse_json_env("{not valid json")


class TestLoadFromFile:
    def test_missing_file_raises(self, tmp_path):
        missing_path = str(tmp_path / "does_not_exist.env")
        with pytest.raises(FileNotFoundError):
            load_from_file(missing_path)

    def test_loads_dotenv_by_extension(self, tmp_path):
        f = tmp_path / ".env.dev"
        f.write_text("KEY=value")
        result = load_from_file(str(f))
        assert result == {"KEY": "value"}

    def test_loads_yaml_by_extension(self, tmp_path):
        f = tmp_path / "config.yaml"
        f.write_text("KEY: value")
        result = load_from_file(str(f))
        assert result == {"KEY": "value"}

    def test_loads_json_by_extension(self, tmp_path):
        f = tmp_path / "config.json"
        f.write_text('{"KEY": "value"}')
        result = load_from_file(str(f))
        assert result == {"KEY": "value"}


class TestLoadConfig:
    def test_fills_default_source_of_truth(self, tmp_path):
        f = tmp_path / "envsync.yaml"
        f.write_text("version: '1'\n")
        config = load_config(str(f))
        assert config["source_of_truth"] == ".env.example"

    def test_fills_default_snapshots(self, tmp_path):
        f = tmp_path / "envsync.yaml"
        f.write_text("version: '1'\n")
        config = load_config(str(f))
        assert config["snapshots"]["directory"] == ".envsync/snapshots"
        assert config["snapshots"]["max_keep"] == 10
        assert config["snapshots"]["encrypted"] is True

    def test_respects_explicit_values(self, tmp_path):
        f = tmp_path / "envsync.yaml"
        f.write_text("source_of_truth: .env.custom\n")
        config = load_config(str(f))
        assert config["source_of_truth"] == ".env.custom"


class TestWriteDotenv:
    def test_writes_simple_values(self, tmp_path):
        path = str(tmp_path / "out.env")
        write_dotenv(path, {"KEY": "value"})
        content = open(path).read()
        assert "KEY=value" in content

    def test_quotes_values_with_spaces(self, tmp_path):
        path = str(tmp_path / "out.env")
        write_dotenv(path, {"NAME": "Vishakha Singh"})
        content = open(path).read()
        assert 'NAME="Vishakha Singh"' in content

    def test_round_trip_dotenv(self, tmp_path):
        path = str(tmp_path / "out.env")
        original = {"HOST": "localhost", "PORT": "5432", "LABEL": "with space"}
        write_dotenv(path, original)
        reloaded = parse_dotenv(open(path).read())
        assert reloaded == original