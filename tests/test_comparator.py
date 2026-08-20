from core.comparator import (
    compare,
    has_drift,
    missing_count,
    mismatch_count,
    extra_count,
    is_sensitive,
    mask_value,
    display_value,
)


class TestCompare:
    def test_matching_keys(self):
        source = {"A": "1"}
        target = {"A": "1"}
        entries = compare(source, target)
        assert len(entries) == 1
        assert entries[0]["status"] == "MATCH"

    def test_missing_key(self):
        source = {"A": "1"}
        target = {}
        entries = compare(source, target)
        assert entries[0]["status"] == "MISSING"
        assert entries[0]["source_value"] == "1"

    def test_extra_key(self):
        source = {}
        target = {"A": "1"}
        entries = compare(source, target)
        assert entries[0]["status"] == "EXTRA"
        assert entries[0]["target_value"] == "1"

    def test_mismatched_value(self):
        source = {"A": "1"}
        target = {"A": "2"}
        entries = compare(source, target)
        assert entries[0]["status"] == "MISMATCH"

    def test_sorted_by_priority_then_key(self):
        source = {"Z_MISSING": "1", "A_MATCH": "1"}
        target = {"A_MATCH": "1", "B_EXTRA": "1"}
        entries = compare(source, target)
        statuses = [e["status"] for e in entries]
        # MISSING(0) < MISMATCH(1) < EXTRA(2) < MATCH(3)
        assert statuses == ["MISSING", "EXTRA", "MATCH"]

    def test_empty_inputs(self):
        assert compare({}, {}) == []


class TestDriftHelpers:
    def test_has_drift_true(self):
        entries = compare({"A": "1"}, {})
        assert has_drift(entries) is True

    def test_has_drift_false(self):
        entries = compare({"A": "1"}, {"A": "1"})
        assert has_drift(entries) is False

    def test_counts(self):
        source = {"MISS": "1", "MISMATCH_KEY": "1", "SAME": "1"}
        target = {"MISMATCH_KEY": "2", "SAME": "1", "EXTRA_KEY": "1"}
        entries = compare(source, target)
        assert missing_count(entries) == 1
        assert mismatch_count(entries) == 1
        assert extra_count(entries) == 1


class TestSensitiveMasking:
    def test_is_sensitive_detects_password(self):
        assert is_sensitive("DB_PASSWORD") is True

    def test_is_sensitive_detects_token(self):
        assert is_sensitive("API_TOKEN") is True

    def test_is_sensitive_false_for_normal_key(self):
        assert is_sensitive("DB_HOST") is False

    def test_mask_short_value(self):
        assert mask_value("ab") == "****"

    def test_mask_long_value_keeps_edges(self):
        masked = mask_value("supersecret123")
        assert masked.startswith("su")
        assert masked.endswith("23")
        assert "*" in masked

    def test_display_value_empty(self):
        assert display_value("ANY_KEY", "") == "(empty)"

    def test_display_value_masks_sensitive(self):
        result = display_value("SECRET_KEY", "mysecretvalue")
        assert "*" in result

    def test_display_value_shows_normal(self):
        assert display_value("DB_HOST", "localhost") == "localhost"