import socket
import logging
import signal
from common.utils import Bet, store_bets
from common.protocol import recv_bet, send_all, send_answer, SUCESS, ERROR_GENERIC

class Server:
    def __init__(self, port, listen_backlog):
        # Initialize server socket
        self._server_socket = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
        self._server_socket.bind(('', port))
        self._server_socket.listen(listen_backlog)
        self.client = None
        self.running = True 

    def run(self):
        """
        Dummy Server loop

        Server that accept a new connections and establishes a
        communication with a client. After client with communucation
        finishes, servers starts to accept new connections again
        """

        signal.signal(signal.SIGTERM, self.__shutdown)
        signal.signal(signal.SIGINT, self.__shutdown)


        # TODO: Modify this program to handle signal to graceful shutdown
        # the server
        while self.running:
            try: 
                client_sock = self.__accept_new_connection()
                if client_sock:
                    self.client = client_sock
                    self.__handle_client_connection()
            except OSError as e:
                if self.running:
                    logging.error(f"action: accept_connections | result: fail | error: {e}")
                break

        logging.info("action: server_shutdown | result: success")


        
    

    def __handle_client_connection(self):
        """
        Read message from a specific client socket and closes the socket

        If a problem arises in the communication with the client, the
        client socket will also be closed
        """
        try:
            bet = recv_bet(self.client)
            store_bets([bet])
            logging.info(f"action: apuesta_almacenada | result: success | dni: {bet.document} | numero: {bet.number}.")
            send_answer(self.client, SUCESS)
        except OSError as e:
            if self.running:
                logging.error(f"action: receive_message | result: fail | error: {e}")
            else:
                logging.info("action: client_shutdown | result: success")
        except Exception: 
            send_answer(self.client, ERROR_GENERIC)
        finally:
            if self.client:
                self.client.close()
            self.client = None


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
        if self.client:
            self.client.close()
            self.client = None
        self.running = False
        self._server_socket.shutdown(socket.SHUT_RDWR)
        self._server_socket.close()
       

       

