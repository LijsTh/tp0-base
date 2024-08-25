INDENT_SPACE = "  "

def parse_dict(d, indent):
    yaml_str = ""
    for key, value in d.items():
        if isinstance(value, dict):
            yaml_str += f"{INDENT_SPACE * indent}{key}:\n" + parse_to_yaml(value, indent + 1)
        elif isinstance(value, list):
            yaml_str += f"{INDENT_SPACE * indent}{key}:\n"
            yaml_str += parse_list(value, indent + 1)
        else:
            yaml_str += f"{INDENT_SPACE * indent}{key}: {value}\n"
    return yaml_str

def parse_list(l, indent):
    yaml_str =""
    for item in l:
        yaml_str += f"{INDENT_SPACE * indent}- {parse_to_yaml(item, indent + 1).lstrip()}"
    return yaml_str


def parse_to_yaml(data , indent=0):
    """
    Parse the data to yaml format with the following conventions:
    dict -> key: value
    list -> - item
    value -> value
    """
    yaml_str = ""
    if isinstance(data, dict):
        yaml_str += parse_dict(data, indent)
    elif isinstance(data, list):
        yaml_str += parse_list(data, indent)
    else:
        yaml_str += f"{INDENT_SPACE * indent}{data}\n"
    return yaml_str
    

def save_yaml_to_file(yaml_str, file_path):
    try:
        with open(file_path, 'w') as file:
            file.write(yaml_str)
    except Exception as e:
        raise e
    


