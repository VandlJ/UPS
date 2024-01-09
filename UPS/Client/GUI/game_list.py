import tkinter as tk
from utils.client_to_server_message import join_msg_creator


# Function to extract game name from game information
def game_name_getter(game_info):
    game_name = game_info.split('-')[0].strip()
    return game_name


# Class responsible for managing the game list screen
class GameListScreen:
    def __init__(self, master, server):
        # Initialize the GameListScreen with master window and server connection
        self.master = master
        self.server = server

        self.game_listbox = None
        self.game_list_screen = None
        
    def game_list_screen_open(self):
        # Open the game list screen and configure its layout
        self.game_list_screen = tk.Toplevel(self.master)
        self.game_list_screen.title("Blackjack")
        self.game_list_screen.protocol("WM_DELETE_WINDOW", self.destroy_master)
        self.game_list_screen.rowconfigure(0, weight=1)
        self.game_list_screen.columnconfigure(0, weight=1)

        self.game_listbox = tk.Listbox(self.game_list_screen)
        self.game_listbox.grid(row=0, column=0, padx=10, pady=10, sticky="nsew")
        self.game_listbox.bind("<Double-Button-1>", self.join)

    def join(self, event):
        # Handle the join event triggered by double-clicking on a game from the list
        game = self.game_listbox.curselection()
        if game:
            selected_item = self.game_listbox.get(game)
            game_name = game_name_getter(selected_item)
            msg = join_msg_creator(game_name)
            self.server.sendall((msg + "\n").encode())

    def update_game_list(self, games):
        # Update the game list based on the information received from the server
        self.game_listbox.delete(0, tk.END)

        for game in games:
            game_name = game['game_name']
            current_players = game['current_players']
            max_players = game['max_players']
            status = 'Waiting' if game['game_status'] == 1 else 'Playing'
            game_info = f"{game_name} - {current_players}/{max_players} ({status})"
            self.game_listbox.insert(tk.END, game_info)

    def withdraw_game_list_screen(self):
        # Withdraw (hide) the game list screen
        self.game_list_screen.withdraw()

    def kill_app2(self):
        # Destroy the game list screen
        self.game_list_screen.destroy()

    def destroy_master(self):
        # Destroy the master window associated with the game list screen
        self.master.destroy()
