import asyncio

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

    #╰( ͡° ͜ʖ ͡° )つ──☆*:・ﾟ write your code here

    # remember to close the connection!
    writer.close()
