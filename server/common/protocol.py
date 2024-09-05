import socket
from common.utils import Bet

## Size constants
AGENCY_SIZE = 1 
STR_SIZE = 1 
DOCUMENT_SIZE = 4 
BIRTHDATE_SIZE = 10 
NUMBER_SIZE = 2 
MAX_STR_SIZE = 255
BATCH_SIZE = 2
ANSWER_SIZE = 1
WINNERS_N_SIZE = 1

## Server answers
SUCESS = 0
ERROR_GENERIC = 1
FINISH = 3

"""
Reads a bet from the socket and returns it.
The bet is expected to be in the following format:
    - 1 byte for the agency
    - 1 byte for the first name size
    - (0-255) bytes for the first name
    - 1 byte for the last name size
    - (0-255) bytes for the last name
    - 4 bytes for the document
    - 10 bytes for the birthdate
    - 2 bytes for the number

"""

def recv_bet(skt: socket.socket) -> Bet:
    data = recv_all(skt, AGENCY_SIZE)
    agency = int.from_bytes(data, byteorder='big')
    data = recv_str_uknown_size(skt)
    first_name = data
    data = recv_str_uknown_size(skt)
    last_name = data
    data = recv_all(skt, DOCUMENT_SIZE)
    document = int.from_bytes(data, byteorder='big')
    data = recv_all(skt, BIRTHDATE_SIZE) 
    birthdate = data.decode('utf-8')

    data = recv_all(skt, NUMBER_SIZE)
    number = int.from_bytes(data, byteorder='big')

    return Bet(agency, first_name, last_name, document, birthdate, number)


def recv_str_uknown_size(skt: socket.socket) -> str:
    data = recv_all(skt, STR_SIZE)
    size = int.from_bytes(data, byteorder='big')
    if size > MAX_STR_SIZE:
        raise ValueError("String size exceeded") 
    data = recv_all(skt, size)
    return data.decode('utf-8')


"""
reads the entire message from the socket to avoid short-reads
"""
def recv_all(skt: socket.socket, size: int) -> bytes:
    data = b'' 
    while len(data) < size:
        pkt = skt.recv(size - len(data))
        if len(pkt) == 0:
            raise OSError("Connection closed")
        data += pkt

    return data

"""
sends the entire message to the socket to avoid short-writes
"""
def send_all(skt: socket.socket, data: bytes) -> None: 
    while len(data) > 0:
        sent = skt.send(data)
        data = data[sent:]

"""
Reads the first byte from the socket to determine the size of the batch
Then proceeds to reach each bet in the batch
"""
def recv_batch(skt: socket.socket) -> list[Bet]:
    data = recv_all(skt, BATCH_SIZE)
    size = int.from_bytes(data, byteorder='big')
    bets = [] 
    for _ in range(size):
        bets.append(recv_bet(skt))
    return bets

def send_error(skt: socket.socket) -> None:
    send_answer(skt, ERROR_GENERIC)

def send_sucess(skt: socket.socket) -> None:
    send_answer(skt, SUCESS)

def send_answer(skt: socket.socket, answer: int) -> None:
    if answer not in [SUCESS, ERROR_GENERIC]:
        raise ValueError("Invalid answer")
    data = answer.to_bytes(ANSWER_SIZE, byteorder='big')
    send_all(skt, data)


def empty_socket(skt: socket.socket) -> None:
    while True:
        try:
            skt.recv(1)
        except:
            ## Socket is empty
            break


def send_results(clients: dict[int, socket.socket], winners: list[int] ) -> None:
    for agency , client in clients.items():
        winner_for_agency = [winner[1] for winner in winners if winner[0] == agency ]
        __send_results(client, winner_for_agency)
        recv_finish(client)

    
def __send_results(client: socket.socket, winners: list[int]) -> None:
    # hago un arreglo de bytes con los ganadores inlcuyendo el tamaño
    data = len(winners).to_bytes(WINNERS_N_SIZE, byteorder='big')
    for winner in winners:
        data += winner.to_bytes(DOCUMENT_SIZE, byteorder='big')
    send_all(client, data)

    
def recv_finish(skt: socket.socket) -> None:
    data = recv_all(skt, ANSWER_SIZE )
    if data[0] != FINISH:
        raise ValueError("Invalid end message")
    

def recv_agency(skt: socket.socket) -> int:
    data = recv_all(skt, AGENCY_SIZE)
    return int.from_bytes(data, byteorder='big')