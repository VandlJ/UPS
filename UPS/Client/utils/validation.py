import re
from tkinter import ttk
import tkinter as tk


def occupied_nick_alert(parent):
    info_message = (
        f"Could not join the lobby! \n\n"
        f"Lobby is in game or full. \n"
    )

    alert_window = tk.Toplevel(parent)
    alert_window.title("Error during joining lobby!")

    label = ttk.Label(alert_window, text=info_message)
    label.pack(padx=10, pady=10)


def disconnected_alert(parent):
    info_message = (
        f"Closing connection! \n\n"
        f"You were disconnect from server.\n"
    )

    alert_window = tk.Toplevel(parent)
    alert_window.title("Closing connection!")

    label = ttk.Label(alert_window, text=info_message)
    label.pack(padx=10, pady=10)


def connection_lost_alert(parent):
    info_message = (
        f"Lost connection with server! \n\n"
        f"You have 60 seconds to join back - if your connection is not gonna be back in next 25 seconds,\n you will be "
        f"disconnected, but still you have 30 seconds to join back with your name to retrieve state.\n"
    )

    alert_window = tk.Toplevel(parent)
    alert_window.title("Connection lost!")

    label = ttk.Label(alert_window, text=info_message)
    label.pack(padx=10, pady=10)


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
