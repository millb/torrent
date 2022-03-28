import hashlib
from typing import Optional

from storage import FileStorage

def calculate_hash(data: bytes) -> str:
    return hashlib.sha1(data).hexdigest()


async def verify_piece(storage: FileStorage, index: int, block_hash: str) -> Optional[bytes]:
    data = await storage.read_block(block=index)

    calculated_hash = calculate_hash(data)
    if block_hash != calculated_hash:
        return None

    return data
