import asyncio
from dataclasses import dataclass
import json
from typing import Optional

CONFIG_PATH = 'torrent.conf'


@dataclass
class FileInfo:
    size: int
    part_size: int
    parts: list[str]

    def block(self, block: int) -> tuple[int, int]:
        """Return offset and size for given block

        :param block: index of block
        :return: (offset, size)
        """

        if block < 0 or block >= len(self.parts):
            raise ValueError('invalid block')

        if block + 1 < len(self.parts):
            return block * self.part_size, self.part_size

        length = self.size % self.part_size
        if length == 0:
            length = self.part_size

        return block * self.part_size, length


@dataclass
class Config:
    peers: list[str]
    file_info: Optional[FileInfo] = None


def read_config(config_path: str = CONFIG_PATH) -> Config:
    with open(config_path, 'r') as f:
        data = json.load(f)

    if not data['FileInfo']['Parts']:
        data['FileInfo']['Parts'] = []
    if not data['Peers']:
        data['Peers'] = []
    return Config(
        peers=data['Peers'],
        file_info=FileInfo(
            size=data['FileInfo']['Size'],
            part_size=data['FileInfo']['PartSize'],
            parts=data['FileInfo']['Parts'],
        ),
    )


class Provider:
    def __init__(self, config_path: str = CONFIG_PATH) -> None:
        self.config = read_config(config_path)
        self.config_path = CONFIG_PATH

        self.lock = asyncio.Lock()

    async def get_config(self) -> Config:
        async with self.lock:
            return self.config
