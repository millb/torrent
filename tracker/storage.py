import threading
import uuid
import copy
import json

class IStorage:
    def GetPeersByHash(self, hash):
        pass

    def CreateUniqueHash(self):
        pass

    def AddNewPeerToHash(self, peer, hash):
        pass

class MemoryStorage(IStorage):
    def __init__(self):
        self.storage = dict()
        self.lock = threading.Lock()

    def GetPeersByHash(self, hash):
        if hash in self.storage:
            return list(self.storage[hash])
        else:
            return list()

    def CreateUniqueHash(self):
        self.lock.acquire()
        new_hash = str(uuid.uuid4())
        while new_hash in self.storage:
            new_hash = uuid4.uuid4()
        self.storage[new_hash] = set()
        self.lock.release()
        return new_hash

    def AddNewPeerToHash(self, peer, hash):
       self.lock.acquire()
       self.storage[hash].add(peer)
       print('!!!!', self.storage)
       self.lock.release()
