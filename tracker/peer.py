from json import JSONEncoder
from json import dumps

class Peer:
    def __init__(self, host, port):
        self.port = port
        self.host = host

    def __str__(self):
        return self.host + ':' + str(self.port)

    def __repr__(self):
        return str(self)

    def __hash__(self):
        print(hash((self.port, self.host)))
        return hash((self.port, self.host))

    def __eq__(self, other):
        return self.port == other.port and self.host == other.host

    def __ne__(self,other):
        return not self.__eq__(other)

class ImprovedEncoder(JSONEncoder):
    def default(self, o):
        return o.__dict__
