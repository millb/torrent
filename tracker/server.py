from flask import Flask, request, jsonify
from storage import MemoryStorage
from peer import Peer, ImprovedEncoder
import logging
import sys

app = Flask(__name__)

logging.basicConfig(level=logging.DEBUG, format=f'[%(asctime)s][%(levelname)s][%(threadName)s]: %(message)s', datefmt='%m/%d/%Y %I:%M:%S')
app.json_encoder = ImprovedEncoder

def get_peer():
    return Peer(request.remote_addr, int(request.json['listening_port']))

@app.route("/ready")
def ready():
    return "I am ready"

@app.route("/peers", methods=['POST'])
def peersHandler():
    app.logger.debug(f'storage:', storage.storage)
    peer = get_peer()
    app.logger.debug(f'peer: {peer}')
    content = request.json
    app.logger.debug(f'Request body: {content}')
    hash = content['hash']
    app.logger.debug(f'hash: {hash}')
    peers = storage.GetPeersByHash(hash)
    app.logger.debug(f'peers: {peers}')
    if len(peers) != 0:
        storage.AddNewPeerToHash(peer, hash)
    return jsonify({'peers': peers})

@app.route("/hash", methods=['POST'])
def hashHandler():
    peer = get_peer()
    app.logger.debug(f'peer: {peer}')
    hash = storage.CreateUniqueHash()
    storage.AddNewPeerToHash(peer, hash)
    app.logger.debug(f'storage:', storage.storage)
    return jsonify({'hash': hash})

if __name__ == "__main__":
    if len(sys.argv) > 1 and sys.argv[1] == 'help':
        print('Usage: python3 server.py address port')
        exit(0)

    storage = MemoryStorage()

    host_name = '0.0.0.0'
    port = 5000
    if len(sys.argv) == 3:
        host_name = sys.argv[1]
        port = int(sys.argv[2])
    app.run(host = host_name, port=port)
