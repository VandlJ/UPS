import tkinter as tk
from tkinter import ttk
from utils.client_to_server_message import create_start_game_message
from utils.client_to_server_message import create_turn_message


class GameWindow:
    def __init__(self, parent, server, chat_window):
        self.start_button_mounted = False
        self.stand_button = None
        self.hit_button = None
        self.hand_value_label = None
        self.cards_in_hand_label = None
        self.nicknames_label = None
        self.current_players_label = None
        self.start_button = None
        self.status_label = None
        self.parent = parent
        self.server = server
        self.chat_window = chat_window
        self.game_window = None
        self._can_be_started = False
        self._current_players = 0
        self._max_players = 0
        self.nicknames = []
        self.points = {}
        self.cards_in_hand = ""
        self.hand_value = ""
        self.game_started = False
        self.game_gui_mounted = False
        self.round_over = False
        self.made_move = False
        self.game_ended = False
        self.winners = []

    @property
    def can_be_started(self):
        return self._can_be_started

    @can_be_started.setter
    def can_be_started(self, value):
        if self._can_be_started != value:
            self._can_be_started = value
            self.refresh_gui()

    @property
    def current_players(self):
        return self._current_players

    @current_players.setter
    def current_players(self, value):
        if self._current_players != value:
            self._current_players = value
            self.refresh_gui()

    @property
    def max_players(self):
        return self._max_players

    @max_players.setter
    def max_players(self, value):
        if self._max_players != value:
            self._max_players = value
            self.refresh_gui()

    def refresh_gui(self):
        print("Can be started: ", self.can_be_started)
        print("Game started: ", self.game_started)
        print("Made move: ", self.made_move)
        if self.game_ended:
            self.show_final_panel()
        elif self.can_be_started and not self.game_started and not self.made_move:
            if not self.start_button_mounted:
                self.start_button.grid(row=1, column=0, pady=10)
                self.start_button_mounted = True
            self.actualize_current_players_label()
            print("BLE1")
        elif self.game_started and not self.made_move:
            self.initialize_game()
            # self.can_be_started = False
            print("BLE2")
        elif self.made_move and not self.game_started:
            self.buttons_update()
            print("BLE3")
        else:
            self.actualize_current_players_label()
            print("BLE4")

    def buttons_update(self):
        print("BUTTONS UPDATE MORE")
        self.hit_button.grid_forget()
        self.stand_button.grid_forget()
        print("BUTTONS UPDATED MORE")

    def initialize_game(self):
        if not self.game_gui_mounted:
            self.status_label.config(text="Game started")

            self.nicknames_label = ttk.Label(self.game_window, text="Players: ")
            self.nicknames_label.grid(row=3, column=0, pady=5)

            self.cards_in_hand_label = ttk.Label(self.game_window, text="Cards: ")
            self.cards_in_hand_label.grid(row=4, column=0, pady=5)

            self.hand_value_label = ttk.Label(self.game_window, text="Hand value: ")
            self.hand_value_label.grid(row=5, column=0, pady=5)

            self.start_button.grid_forget()

            self.nicknames_label.config(text=f"Player: {self.nicknames[0]}")
            self.cards_in_hand_label.config(text=f"Cards in hand: {self.cards_in_hand}")
            self.hand_value_label.config(text=f"Hand value: {self.hand_value}")

            self.game_gui_mounted = True

        players_info = f"{self._current_players}/{self._max_players}"
        self.current_players_label.config(text=f"Current players: {players_info}")

        self.nicknames_label.config(text=f"Player: {self.nicknames[0]}")
        self.cards_in_hand_label.config(text=f"Cards in hand: {self.cards_in_hand}")
        self.hand_value_label.config(text=f"Hand value: {self.hand_value}")

        self.hit_button = tk.Button(self.game_window, text="Hit", command=lambda: self.send_move("HIT"))
        self.stand_button = tk.Button(self.game_window, text="Stand", command=lambda: self.send_move("STAND"))

        self.hit_button.grid(row=6, column=0, pady=10)
        self.stand_button.grid(row=7, column=0, pady=10)

    def actualize_current_players_label(self):
        players_info = f"{self._current_players}/{self._max_players}"
        self.current_players_label.config(text=f"Current players: {players_info}")

    def open_game_window(self):
        self.game_window = tk.Toplevel(self.parent)
        self.game_window.title("Blackjack Game Window")

        self.status_label = tk.Label(self.game_window, text="Waiting for players")
        self.status_label.grid(row=0, column=0, pady=10)

        self.start_button = tk.Button(self.game_window, text="Start the Game", command=self.start_game)

        self.current_players_label = tk.Label(self.game_window, text="Current players: ")
        self.current_players_label.grid(row=2, column=0, pady=5)

        self.refresh_gui()

    def send_move(self, turn):
        message = create_turn_message(turn)
        self.server.sendall((message + "\n").encode())
        self.made_move = True
        self.game_started = False
        self.refresh_gui()

    def start_game(self):
        message = create_start_game_message()
        self.server.sendall((message + "\n").encode())

    def segment_handler(self, message_body):
        segments = message_body.split('|')

        if len(segments) == 3:
            player_info = segments[0].split(';')
            self.nicknames = []
            self.points = {}

            for player_data in player_info:
                player_data_parts = player_data.split(':')
                if len(player_data_parts) == 2:
                    nickname, points_str = player_data_parts
                    self.nicknames.append(nickname)
                    self.points[nickname] = int(points_str)

            self.cards_in_hand = segments[1]
            self.hand_value = segments[2]
            self.refresh_gui()

            print("SEGMENT - Can be started: ", self.can_be_started)
            print("SEGMENT - Game started: ", self.game_started)
            print("SEGMENT - Made move: ", self.made_move)

        else:
            return None

    def extract_turn_info(self, message_body):
        self.game_started = False
        self.segment_handler(message_body)

    def extract_init_game_info(self, message_body):
        self.game_started = True
        self.segment_handler(message_body)

    def end_the_game(self, message_body):
        nicknames = message_body.split(';')
        self.winners = nicknames
        self.game_ended = True
        self.refresh_gui()

    def show_final_panel(self):
        if len(self.winners) > 1:
            self.cards_in_hand_label.config(text="Winners: " + ";".join(self.winners))
        else:
            self.cards_in_hand_label.config(text="Winner: " + self.winners[0])
        self.status_label.config(text="Game over")
        self.hand_value_label.config(text="Soon you will be moved beck to the main lobby!")
        self.nicknames_label.config(text="")
        self.game_window.after(10000, self.close_window)

    def close_window(self):
        self.game_window.destroy()
        self.chat_window.deiconify()
