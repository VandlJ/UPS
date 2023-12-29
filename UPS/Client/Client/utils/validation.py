import re


def validate_nickname(nickname):
    pattern = re.compile(r'^[a-zA-Z0-9_]+$')
    return bool(pattern.match(nickname))


def validate_server_ip(server_ip):
    pattern = re.compile(r'^\d{1,3}\.\d{1,3}\.\d{1,3}\.\d{1,3}$')

    if pattern.match(server_ip) or server_ip.lower() == 'localhost':
        return True
    else:
        return False


def validate_server_port(server_port):
    pattern = re.compile(r'^[1-9]\d*$')
    server_port_string = str(server_port)
    return bool(pattern.match(server_port_string))
