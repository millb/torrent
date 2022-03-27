import asyncio

from config import Provider
from storage import FileStorage
from solution import handle_socket

port = # choose any port you like

async def main() -> None:
    #╰( ͡° ͜ʖ ͡° )つ──☆*:・ﾟ write your code here

    # create server
    server = await asyncio.start_server(handle_socket, '0.0.0.0', port)
    addrs = ', '.join(str(sock.getsockname()) for sock in server.sockets)
    print(f'Serving on {addrs}', flush=True)

    #╰( ͡° ͜ʖ ͡° )つ──☆*:・ﾟ write your code here

    await asyncio.gather(
        server.serve_forever(),
        background_routine(),
    )

async def background_routine() -> None:
    #╰( ͡° ͜ʖ ͡° )つ──☆*:・ﾟ write your code here

    # remember to use asyncio.sleep() and NOT time.sleep() for sleeping:
    # await asyncio.sleep(1)

    #╰( ͡° ͜ʖ ͡° )つ──☆*:・ﾟ write your code here
    # for example, you can use this function to connect to other peers

asyncio.run(main())
