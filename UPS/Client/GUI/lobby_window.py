import tkinter as tk
from utils.client_to_server_message import create_game_joining_message


def extract_game_name(game_info):
    game_name = game_info.split('-')[0].strip()
    return game_name


class LobbyWindow:
    def __init__(self, parent, server):
        self.parent = parent
        self.server = server
        self.game_listbox = None
        self.chat_window = None

    def open_chat_window(self):
        self.chat_window = tk.Toplevel(self.parent)
        self.chat_window.title("Blackjack Lobby Window")

        self.game_listbox = tk.Listbox(self.chat_window)
        self.game_listbox.grid(row=0, column=0, padx=10, pady=10, sticky="nsew")
        self.game_listbox.bind("<Double-Button-1>", self.on_double_click)
        self.chat_window.rowconfigure(0, weight=1)
        self.chat_window.columnconfigure(0, weight=1)

    def on_double_click(self, event):
        selected_index = self.game_listbox.curselection()
        if selected_index:
            selected_item = self.game_listbox.get(selected_index)
            game_name = extract_game_name(selected_item)
            message = create_game_joining_message(game_name)
            self.server.sendall((message + "\n").encode())

    def update_game_list(self, games):
        self.game_listbox.delete(0, tk.END)

        for game in games:
            game_name = game['game_name']
            current_players = game['current_players']
            max_players = game['max_players']
            status = 'Waiting' if game['game_status'] == 1 else 'Playing'
            game_info = f"{game_name} - {current_players}/{max_players} ({status})"
            self.game_listbox.insert(tk.END, game_info)

    def close_lobby_window(self):
        self.chat_window.withdraw()
