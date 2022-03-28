'''
BAD TESTS
'''

import json
import requests
import time

host_name = 'http://localhost:5000/'

print('### WAIT READINESS ###')
is_ok = False
for i in range(60):
    r = requests.get(host_name + 'ready')
    if r.status_code == 200:
        is_ok = True
        break
    time.sleep(1)

if not is_ok:
    print('ERROR')
    exit(0)

print('### TEST HASH 1 ###')
r = requests.post(host_name + 'hash')
hash1 = r.json()['hash']
print(hash1)

print('### TEST HASH 2 ###')
r = requests.post(host_name + 'hash')
hash2 = r.json()['hash']
print(hash2)

print('### TEST PEERS 1.1 ###')
r = requests.post(host_name + 'peers', json={'hash': hash1})
print(r.json())

print('### TEST PEERS 1.2 ###')
r = requests.post(host_name + 'peers', json={'hash': hash1})
print(r.json())

print('### TEST PEERS 2.1 ###')
r = requests.post(host_name + 'peers', json={'hash': hash2})
print(r.json())

print('### TEST PEERS 1.3 ###')
r = requests.post(host_name + 'peers', json={'hash': hash1})
print(r.json())

print('### TEST PEERS 2.2 ###')
r = requests.post(host_name + 'peers', json={'hash': hash2})
print(r.json())