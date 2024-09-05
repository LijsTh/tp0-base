import socket
import logging
import signal
import multiprocessing
from common.utils import store_bets, load_bets, has_won
from common.protocol import  recv_batch, send_error, send_sucess,recv_agency, send_results 

MAX_AGENCIES = 5
class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.running = True 
        self.pool = multiprocessing.Pool(MAX_AGENCIES)

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. The server will launch a new
        process to handle the client connection. 
        """

        signal.signal(signal.SIGTERM, self.__shutdown)
        signal.signal(signal.SIGINT, self.__shutdown)


        manager = multiprocessing.Manager()
        file_lock = manager.Lock()
        barrier = manager.Barrier(MAX_AGENCIES)

        with manager : 
            while self.running:
                try: 

                    client_sock = self.__accept_new_connection()

                    if client_sock:
                        self.pool.apply_async(handle_client, (client_sock, file_lock, barrier))

                except OSError as e:
                    if self.running:
                        logging.error(f"action: accept_connections | result: fail | error: {e}")
                    break

        # wait for all the handlers to finish
        self.pool.close()
        self.pool.join()

        logging.info("action: server_shutdown | result: success")



    def __accept_new_connection(self):
        """
        Accept new connections

        Function blocks until a connection to a client is made.
        Then connection created is printed and returned
        """

        # Connection arrived
        logging.info('action: accept_connections | result: in_progress')
        try:
            c, addr = self._server_socket.accept()
            logging.info(f'action: accept_connections | result: success | ip: {addr[0]}')
            return c
        except OSError or KeyboardInterrupt as e:
            if self.running:
                raise e

    
    def __shutdown(self, signum, frame):
        self.running = False
        self._server_socket.shutdown(socket.SHUT_RDWR)
        self._server_socket.close()


def handle_client(client, file_lock,  barrier):
    try:
        bets = recv_batch(client)
        if len(bets) == 0:
            agency = recv_agency(client)
            last_one = barrier.n_waiting == MAX_AGENCIES - 1
            barrier.wait()

            # Sincronized so can read the file
            bets = list(load_bets())

            winners = [int(bet.document) for bet in bets if has_won(bet) and bet.agency == agency]
            if last_one:
                logging.info(f"action: sorteo | result: success")
            send_results(client, winners)
                
        else :
            with file_lock:
                store_bets(bets)
            logging.info(f"action: apuesta_recibida | result: success | cantidad: {len(bets)}")
            send_sucess(client)

    except OSError as e:  # Connection closed
        return

    except Exception as e :
            logging.info(f"apuesta_recibida | result: fail | error: {e}")
            send_error(client) 

    finally:
        client.close()