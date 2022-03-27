import asyncio
import socket
import json
import random

from config import read_config, Config
from storage import FileStorage
from solution import handle_socket
from utils import verify_piece

port = 1000# choose any port you like

def UploadFile(file_path, torrent_path):
    # Create torrent file from file on path.
    # Segment -> hash
    # ToTracker -> Tracker::CreateHash()

def DownloadFile(torrent_path, file_path):
    # Tracker::GetFile() by hash_func

async def main() -> None:
    #╰( ͡° ͜ʖ ͡° )つ──☆*:・ﾟ write your code here

    cfg = read_config()
    server = await asyncio.start_server(handle_socket, '0.0.0.0', port)
    addrs = ', '.join(str(sock.getsockname()) for sock in server.sockets)
    print(f'Serving on {addrs}', flush=True)
    print(f'file_info: size: {cfg.file_info.size}, part_size: {cfg.file_info.part_size}', flush=True)

    await asyncio.gather(
        server.serve_forever(),
        background_routine(cfg),
    )


async def background_routine(cfg) -> None:
    await asyncio.sleep(5)
    name = socket.gethostname()
    ip = socket.gethostbyname(name)
    print(f'name: {name}, ip: {ip}', flush=True)

    file_storage = FileStorage(cfg.file_info)

    random_parts = [(i, cfg.file_info.parts[i]) for i in range(len(cfg.file_info.parts))]
    random.shuffle(random_parts)

    for idx, part in random_parts:
        if await verify_piece(file_storage, idx, cfg.file_info.parts[idx]):
            print(f'Have {idx}')
            await request_about_node(idx, file_storage, ip, cfg)


async def request_about_node(block_idx, file_storage, ip, cfg):
    data = await file_storage.read_block(block_idx)
    print(f'data read {len(data)}, {block_idx}', flush=True)
    block_size = cfg.file_info.block(block_idx)[1]
    for peer in cfg.peers:
        await send_look_block(peer, block_idx, data, block_size, ip, cfg)


async def send_look_block(peer, block_index, data, block_size, ip, cfg):
    if peer == ip:
        return

    print(f'Trying to connect to {peer}', flush=True)
    reader, writer = await asyncio.open_connection(peer, 1000)
    print(f'Connect to {peer}', flush=True)

    request = 'look-' + str(block_index)
    writer.write(request.encode())
    await writer.drain()

    resp = await reader.read(128)
    if resp.decode('utf-8') != 'Nope':
        print(f'send to {peer} {block_index}', flush=True)
        offset = 0
        while offset < block_size:
            writer.write(data[offset:offset+min(128, block_size-offset)])
            await writer.drain()
            # result = await reader.read(128)
            # while (block_index + 1) * cfg.file_info.part_size < cfg.file_info.size and result.decode('utf-8') == 'fail':
            #     print(f'Again for {block_index}')
            #     writer.write(data[offset:offset + min(128, block_size - offset)])
            #     await writer.drain()
            #     result = await reader.read(128)
            offset += 128
    writer.close()


asyncio.run(main())
