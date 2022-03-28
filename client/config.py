import asyncio
from dataclasses import dataclass
import json
from typing import Optional, List, Tuple

@dataclass
class FileInfo:
    hash: str
    size: int
    part_size: int
    parts: List[str]

    def block(self, block: int) -> Tuple[int, int]:
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
    peers: List[str]
    file_info: Optional[FileInfo] = None

CONFIG_PATH = 'torrent.conf'
def read_config(config_path: str = CONFIG_PATH) -> Config:
    with open(config_path, 'r') as f:
        data = json.load(f)

    if not data['FileInfo']['Parts']:
        data['FileInfo']['Parts'] = []
    if 'Peers' not in data:
        data['Peers'] = []
    return Config(
        peers=data['Peers'],
        file_info=FileInfo(
            size=data['FileInfo']['Size'],
            part_size=data['FileInfo']['PartSize'],
            parts=data['FileInfo']['Parts'],
            hash=data['FileInfo']['Hash'],
        ),
    )
