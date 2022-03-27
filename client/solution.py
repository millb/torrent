import asyncio
from config import read_config, Config
from storage import FileStorage
from utils import verify_piece
from typing import Optional
import json

async def handle_socket(reader: asyncio.StreamReader, writer: asyncio.StreamWriter) -> None:
    # this code should handle the connection
    # this is a P2P network, and there is likely no separation between server and client
    #
    # writer can be used to send data to the peer:
    # writer.write(some_bytes)
    # await writer.drain()
    #
    # reader can be used to read data from the peer:
    # resp = await reader.read(length)
    config = read_config()

    file_storage = FileStorage(config.file_info)
    request = await reader.read(128)
    print(f'Got {request.decode("utf-8")}', flush=True)
    notification, block_idx = request.decode('utf-8').split('-')
    block_idx = int(block_idx)

    if notification == 'look':
        dont_want_take = await verify_piece(file_storage, block_idx, config.file_info.parts[block_idx])
        if dont_want_take:
            response = 'Nope'.encode()
            writer.write(response)
            await writer.drain()
        else:
            response = 'ok'.encode()
            writer.write(response)
            await writer.drain()

            result = bytes()
            offset = 0
            block_size = config.file_info.block(block_idx)[1]
            while offset < block_size:
                mini_block = await reader.read(min(128, block_size-offset))
                # while (block_idx + 1) * config.file_info.part_size < config.file_info.size and len(mini_block) < 128:
                #     print(f'Again for {block_idx}')
                #     writer.write('fail'.encode())
                #     await writer.drain()
                #     mini_block = await reader.read(min(128, block_size-offset))
                #
                # writer.write('good'.encode())
                # await writer.drain()

                result += mini_block
                offset += 128

            if block_size % 128 == 0 and len(mini_block) < 128:
                print(f'Strange situation with {block_idx}')

            await file_storage.write_block(block_idx, result)
            if not await verify_piece(file_storage, block_idx, config.file_info.parts[block_idx]):
                print(f"Wtf {block_idx}, {block_size}, {block_num}", flush=True)
            print(f'Successfully got {block_idx}', flush=True)

    #╰( ͡° ͜ʖ ͡° )つ──☆*:・ﾟ write your code here

    # remember to close the connection!
    writer.close()
