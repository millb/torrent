import asyncio
import pathlib
from typing import Optional

from config import FileInfo

DATA_PATH = 'data.bin'


class FileStorage:
    def __init__(self, info: FileInfo, path: str = DATA_PATH) -> None:
        p = pathlib.Path(path)
        p.touch(exist_ok=True)

        self.path = path
        self.file_config: FileInfo = info
        # Open for read and write
        self.file = p.open('rb+')
        self.file.truncate(self.file_config.size)
        self.lock = asyncio.Lock()

    async def read_block(self, block: int) -> bytes:
        async with self.lock:
            offset, size = self.file_config.block(block=block)
            try:
                self.file.seek(offset, 0)
            except Exception as any_exc:
                print('error while file seeking', flush=True)
                raise any_exc

            buf = self.file.read(size)

        return buf

    async def write_block(self, block: int, data: bytes) -> None:
        async with self.lock:
            offset, size = self.file_config.block(block=block)

            if len(data) != size:
                raise ValueError('invalid data size')

            try:
                self.file.seek(offset, 0)
            except Exception as any_exc:
                print('error while file seeking', flush=True)
                raise any_exc

            self.file.write(data)
