from yaml_parsing import parse_to_yaml, save_yaml_to_file

def clients_data(clients_n):
    clients = {}
    for i in range(clients_n):
        clients[f"client{i}"] = {
            "container_name": f"client{i}",
            "image": "client:latest",
            "entrypoint": "/client",
            "environment": [
                f'CLI_ID={i}',
                "CLI_LOG_LEVEL=DEBUG"
            ],
            "networks": ["testing_net"],
            "depends_on": ["server"]
        }
    return clients


def server_data():
    return {
        "container_name": "server",
        "image": "server:latest",
        "entrypoint": "python3 /main.py",
        "environment": [
            "PYTHONUNBUFFERED=1",
            "LOGGING_LEVEL=DEBUG"
        ],
        "networks": ["testing_net"]
    }

def network_data():
    return {
        "testing_net": {
            "ipam" : {
                "driver": "default",
                "config" : [{
                    "subnet": "172.25.125.0/24"
                }]
            }
        }
    }


def write_docker_compose_file(data):
    yaml_data = {
        "name" : "tp0",
        "services": {
            "server": server_data(),
            **clients_data(data["clients_n"]),
        },
        "networks": network_data()
    }

    yaml_str = parse_to_yaml(yaml_data)
    save_yaml_to_file(yaml_str,"docker-compose-test.yml")



            


    

if __name__ == "__main__":
    data = {
        "clients_n": 3
    }
    write_docker_compose_file(data)