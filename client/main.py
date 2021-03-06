import asyncio
import socket
import json
import random
import requests
import os.path
import os
import uuid
import sys
import shutil
import socket

from config import read_config, Config, FileInfo
from storage import FileStorage
from utils import verify_piece
from utils import calculate_hash

CONFIG_PATH = './cfg.json'
PART_SIZE = 256

def hash_to_torrent_path(torrent_dir, hash):
    return torrent_dir + hash + '.json'

def hash_to_file_path(file_dir, hash):
    return file_dir + hash + '.bin'

class Manager:
    def __init__(self, cfg_path: str = CONFIG_PATH):
        print(cfg_path)
        with open(cfg_path, 'r') as f:
            self.cfg = json.load(f)
            self.port = self.cfg['self']['port']
            self.host = self.cfg['self']['host']
            self.name = self.cfg['self']['service']
            self.tracker_host = self.cfg['tracker']['host']
            self.tracker_port = self.cfg['tracker']['port']
            self.files_dir = self.name + '/files/'
            self.torrent_dir = self.name + '/torrents/'
            os.makedirs(self.files_dir, exist_ok=True)
            os.makedirs(self.torrent_dir, exist_ok=True)
        self.hashes = set()
        for filename in os.listdir(self.torrent_dir):
            self.hashes.add(filename[:-5])
        print('*** HASHES ***')
        for hash in self.hashes:
            print(hash, flush=True)
        print('**************')

    async def _uploadFile(self, file_path, path_to_end_file):
        req_path = self.tracker_host + ':' + str(self.tracker_port) + '/hash'
        response = requests.post(req_path, json={'listening_port': self.port, 'listening_host': self.host})
        json_response = response.json()
        hash = json_response['hash']
        torrent_path = await self._createTorrent(file_path, hash)
        real_file_path = self.files_dir + '/' + hash + '.bin'
        shutil.copy(file_path, real_file_path)
        # os.makedirs(path_to_end_file, exist_ok=True)
        shutil.copy(torrent_path, path_to_end_file)
        self.hashes.add(hash)
        print('SUCCESSFUL UPLOAD', flush=True)


    async def _downloadFile(self, torrent_path, path_to_end_file):
        with open(torrent_path, 'r') as f:
            data = json.load(f)
        path = self.tracker_host + ':' + str(self.tracker_port) + '/peers'
        hash = data['FileInfo']['Hash']
        response = requests.post(path, json={'hash': hash, 'listening_port': self.port, 'listening_host': self.host})
        data['Peers'] = response.json()
        cfg = Config(
            peers=data['Peers'],
            file_info=FileInfo(
                size=data['FileInfo']['Size'],
                part_size=data['FileInfo']['PartSize'],
                parts=data['FileInfo']['Parts'],
                hash=data['FileInfo']['Hash'],
            ),
        )
        shutil.copy(torrent_path, hash_to_torrent_path(self.torrent_dir, hash))
        self.hashes.add(hash)
        await self._background_routine(cfg, hash_to_file_path(self.files_dir, hash))
        # os.makedirs(path_to_end_file, exist_ok=True)
        shutil.copy(hash_to_file_path(self.files_dir, hash), path_to_end_file)
        print('SUCCESSFUL DOWNLOAD', flush=True)


    async def startServer(self) -> None:
        print(self.host, self.port)
        print(type(self.host), type(self.port))
        server = await asyncio.start_server(self._handle_socket, self.host, self.port)
        addrs = ', '.join(str(sock.getsockname()) for sock in server.sockets)
        print(f'Serving on {addrs}', flush=True)
        await asyncio.gather(
            server.serve_forever(),
            self._listen_client(),
        )


    async def _createTorrent(self, file_path: str, hash: str, part_size: int = PART_SIZE):
        with open(file_path, 'rb') as file:
            size = os.path.getsize(file_path)
            current_ptr = 0

            file_info = {
                'FileInfo': {
                    'Size': size,
                    'PartSize': part_size,
                    'Parts': [],
                    'Hash': hash,
                }
            }
            while current_ptr < size:
                bytes = file.read(min(part_size, size-current_ptr))
                file_info['FileInfo']['Parts'].append(calculate_hash(bytes))
                current_ptr += part_size

            torrent_path = self.torrent_dir + hash + '.json'
            f = open(torrent_path, 'w')
            f.write(json.dumps(file_info))
            f.close()
        return torrent_path


    async def _listen_client(self):
        loop = asyncio.get_event_loop()
        while True:
            cmd = await loop.run_in_executor(None, input, '')
            command = cmd.split()
            if command[0] == 'download':
                await self._downloadFile(command[1], command[2])
            elif command[0] == 'upload':
                await self._uploadFile(command[1], command[2])


    async def _background_routine(self, cfg, file_path) -> None:
        await asyncio.sleep(5)
        print(f'ip: {self.host}', flush=True)

        file_storage = FileStorage(cfg.file_info, file_path)

        random_parts = [(i, cfg.file_info.parts[i]) for i in range(len(cfg.file_info.parts))]
        random.shuffle(random_parts)

        for idx, part in random_parts:
#         for idx in range(len(cfg.file_info.parts)):
            print(idx)
            print(cfg.peers)
            for obj in cfg.peers['peers']:
                peer = obj['host']
                port = obj['port']
                await self._request_about_node(peer, port, idx, file_storage, cfg)


    async def _request_about_node(self, peer, port, block_idx, file_storage, cfg):
        if peer == self.host and port == self.port:
            return

        block = await verify_piece(file_storage, block_idx, cfg.file_info.parts[block_idx])
        if block:
            return

        print(f'Trying to connect to {peer}:{port}', flush=True)
        reader, writer = await asyncio.open_connection(peer, port)
        print(f'Connect to {peer}:{port}', flush=True)

        request = 'need ' + str(cfg.file_info.hash) + ' ' + str(block_idx)
        writer.write(request.encode())
        await writer.drain()

        resp = await reader.read(8) # keyword (nononono or gogogogo)

        print(resp.decode('ascii'))
        if resp.decode('ascii') != 'nononono':
            print(f'get from {peer}: {block_idx}', flush=True)
            result = bytes()
            offset = 0
            block_size = cfg.file_info.block(block_idx)[1]
            while offset < block_size:
                mini_block = await reader.read(min(256, block_size - offset))
                result += mini_block
                offset += 256
            await file_storage.write_block(block_idx, result)
            print(f'Successfully got {block_idx}', flush=True)
        writer.close()


    async def _handle_socket(self, reader: asyncio.StreamReader, writer: asyncio.StreamWriter) -> None:
        request = await reader.read(256)
        req = request.decode('utf-8')
        print(f'Got {request.decode("utf-8")}', flush=True)
        verification_word, hash, block_index = req.split()
        block_index = int(block_index)
        if verification_word != 'need':
            writer.close()
            return

        if hash not in self.hashes:
            print('Nope hash')
            writer.write('nononono'.encode('ascii'))
            await writer.drain()
            writer.close()
            return

        cfg = read_config(hash_to_torrent_path(self.torrent_dir, hash))
        file_storage = FileStorage(cfg.file_info, hash_to_file_path(self.files_dir, hash))
        block = await verify_piece(file_storage, block_index, cfg.file_info.parts[block_index])
        if not block:
            print(' bock')
            writer.write('nononono'.encode('ascii'))
            await writer.drain()
            writer.close()
            return

        writer.write('gogogogo'.encode('ascii'))
        print('gogogogo')
        await writer.drain()

        writer.write(block)
        await writer.drain()
        writer.close()


asyncio.run(Manager(sys.argv[1]).startServer())