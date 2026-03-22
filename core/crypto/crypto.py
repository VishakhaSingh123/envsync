import os
import base64
from cryptography.hazmat.primitives.ciphers.aead import AESGCM
from cryptography.hazmat.primitives import hashes
from cryptography.hazmat.backends import default_backend


def derive_key(passphrase):
    digest = hashes.Hash(hashes.SHA256(), backend=default_backend())
    digest.update(passphrase.encode())
    return digest.finalize()


def encrypt(plaintext, passphrase):
    key = derive_key(passphrase)
    aesgcm = AESGCM(key)
    nonce = os.urandom(12)
    ciphertext = aesgcm.encrypt(nonce, plaintext.encode(), None)
    return base64.b64encode(nonce + ciphertext).decode()


def decrypt(encoded, passphrase):
    key = derive_key(passphrase)
    data = base64.b64decode(encoded)
    nonce = data[:12]
    ciphertext = data[12:]
    aesgcm = AESGCM(key)
    plaintext = aesgcm.decrypt(nonce, ciphertext, None)
    return plaintext.decode()


def encrypt_map(kv, passphrase):
    return {k: encrypt(v, passphrase) for k, v in kv.items()}


def decrypt_map(kv, passphrase):
    return {k: decrypt(v, passphrase) for k, v in kv.items()}


def get_encryption_key(env_var_name):
    key = os.environ.get(env_var_name)
    if not key:
        raise ValueError(
            f"Encryption key not found. Set it with:\n"
            f"  export {env_var_name}=$(python -c \"import secrets; print(secrets.token_urlsafe(32))\")"
        )
    return key


def generate_key():
    import secrets
    return secrets.token_urlsafe(32)
