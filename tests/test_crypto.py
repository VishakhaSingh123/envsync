import pytest
from core.crypto import encrypt, decrypt, encrypt_map, decrypt_map, derive_key, generate_key


class TestEncryptDecrypt:
    def test_round_trip(self):
        original = "super-secret-value"
        passphrase = "my-passphrase"
        encrypted = encrypt(original, passphrase)
        decrypted = decrypt(encrypted, passphrase)
        assert decrypted == original

    def test_encrypted_value_differs_from_plaintext(self):
        encrypted = encrypt("hello", "passphrase")
        assert encrypted != "hello"

    def test_wrong_passphrase_fails_to_decrypt(self):
        encrypted = encrypt("hello", "correct-pass")
        with pytest.raises(Exception):
            decrypt(encrypted, "wrong-pass")

    def test_encrypting_same_value_twice_differs(self):
        # nonce is random each time, so ciphertext should not be identical
        e1 = encrypt("hello", "passphrase")
        e2 = encrypt("hello", "passphrase")
        assert e1 != e2

    def test_empty_string_round_trip(self):
        encrypted = encrypt("", "passphrase")
        assert decrypt(encrypted, "passphrase") == ""


class TestEncryptDecryptMap:
    def test_round_trip_map(self):
        kv = {"DB_PASSWORD": "hunter2", "API_KEY": "abc123"}
        passphrase = "my-passphrase"
        encrypted = encrypt_map(kv, passphrase)
        decrypted = decrypt_map(encrypted, passphrase)
        assert decrypted == kv

    def test_empty_map(self):
        assert encrypt_map({}, "passphrase") == {}
        assert decrypt_map({}, "passphrase") == {}


class TestDeriveKey:
    def test_deterministic(self):
        assert derive_key("same-pass") == derive_key("same-pass")

    def test_different_passphrases_differ(self):
        assert derive_key("pass1") != derive_key("pass2")

    def test_key_length_is_32_bytes(self):
        # AES-256 requires a 32-byte key
        assert len(derive_key("any-pass")) == 32


class TestGenerateKey:
    def test_generates_nonempty_string(self):
        key = generate_key()
        assert isinstance(key, str)
        assert len(key) > 0

    def test_generates_unique_keys(self):
        assert generate_key() != generate_key()