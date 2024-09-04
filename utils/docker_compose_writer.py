import sys

N_CLIENTS = 2
FILE_NAME = 1
N_ARGS = 3

def clients_data(clients_n):
    client_total = ""
    for i in range(1, clients_n+1):
        client_total += f"""
  client{i}:
     container_name: client{i}
     image: client:latest
     entrypoint: /client
     networks:
       - testing_net
     depends_on:
       - server 
     volumes:
       - ./client/config.yaml:/config.yaml
       - ./.data/:/data\n"""
    return client_total




def server_data():
    return f"""
  server:
    container_name: server
    image: server:latest
    entrypoint: python3 /main.py
    networks:
      - testing_net
    volumes:
      - ./server/config.ini:/config.ini\n"""


def network_data():
    return f"""
networks:
  testing_net:
    ipam:
      driver: default
      config:
        - subnet: 172.25.125.0/24\n"""



def write_docker_compose_file(file_name,  n_clients):
    yaml_str = f"""name: tp0\nservices:"""
    yaml_str += server_data()
    yaml_str += clients_data(n_clients)
    yaml_str += network_data()

    with open(file_name, "w") as f:
        f.write(yaml_str)




if __name__ == "__main__":
    try:
        if len(sys.argv) != N_ARGS:
            raise ValueError("Uso: python3 docker_compose_writer.py <nombre_del_archivo_docker_compose> <n_clientes>")
        elif not sys.argv[N_CLIENTS].isdigit() or int(sys.argv[N_CLIENTS]) <= 0:
            raise ValueError("n_clientes debe ser un número entero positivo")
        elif not sys.argv[FILE_NAME].endswith(".yaml"):
            raise ValueError("El nombre del archivo debe terminar con .yaml")
            
        write_docker_compose_file(sys.argv[FILE_NAME], int(sys.argv[N_CLIENTS]))
        print(f"Archivo {sys.argv[FILE_NAME]} generado correctamente!")

    except Exception as e:
        print(f'Error al general el docker-compose : {e}')
        sys.exit(1)
