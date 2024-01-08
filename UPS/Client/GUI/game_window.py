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
        self.disconnected_nicknames = []
        self.points = {}
        self.standing_players = []
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
        if self.game_ended:
            self.show_final_panel()
        elif self.can_be_started and not self.game_started and not self.made_move:
            if not self.start_button_mounted:
                self.start_button.grid(row=1, column=0, pady=10)
                self.start_button_mounted = True
            self.actualize_current_players_label()
        elif self.game_started and not self.made_move:
            self.initialize_game()
            # self.can_be_started = False
        elif self.made_move and not self.game_started:
            self.buttons_update()
        else:
            self.actualize_current_players_label()

    def buttons_update(self):
        print("BUTTONS FORGET")
        self.hit_button.grid_forget()
        self.stand_button.grid_forget()

    def destroy_parent(self):
        self.parent.destroy()

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

            connected_players = [f"{nickname}" for nickname in self.nicknames]
            disconnected_players = [f"[Lost Connection] {nickname}" for nickname in self.disconnected_nicknames]
            players_string = ", ".join(connected_players + disconnected_players)

            self.nicknames_label.config(text=f"Players: {players_string}")
            self.cards_in_hand_label.config(text=f"Cards in hand: {self.cards_in_hand}")
            self.hand_value_label.config(text=f"Hand value: {self.hand_value}")

            self.game_gui_mounted = True

        players_info = f"{self._current_players}/{self._max_players}"
        self.current_players_label.config(text=f"Current players: {players_info}")

        connected_players = [f"{nickname}" for nickname in self.nicknames]
        disconnected_players = [f"[Lost Connection] {nickname}" for nickname in self.disconnected_nicknames]
        players_string = ", ".join(connected_players + disconnected_players)

        self.nicknames_label.config(text=f"Player: {players_string}")
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
        self.game_window.protocol("WM_DELETE_WINDOW", self.destroy_parent)

        self.status_label = tk.Label(self.game_window, text="Waiting for players")
        self.status_label.grid(row=0, column=0, pady=10)

        self.start_button = tk.Button(self.game_window, text="Start the Game", command=self.start_game)

        self.current_players_label = tk.Label(self.game_window, text="Current players: ")
        self.current_players_label.grid(row=2, column=0, pady=5)

        self.refresh_gui()

    def send_move(self, turn):
        if turn == "STAND":
            self.standing_players.append(self.nicknames[0])

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
        print("BBBBBBBBBB:", segments)
        if len(segments) == 3:
            self.segment_splitter(segments)
            self.refresh_gui()
        elif len(segments) == 4:
            print("4 SEGMENTS")
            self.segment_splitter(segments)
            if segments[3] == "1":
                print("TU SOM")
                self.refresh_gui()
                self.buttons_update()
            else:
                self.refresh_gui()
        else:
            return None

    def segment_splitter(self, segments):
        player_info = segments[0].split(';')
        self.nicknames = []

        for player_data in player_info:
            nickname = player_data
            self.nicknames.append(nickname)

        self.cards_in_hand = segments[1]
        self.hand_value = segments[2]

    def extract_turn_info(self, message_body):
        self.game_started = False
        self.segment_handler(message_body)

    def extract_init_game_info(self, message_body):
        self.game_started = True
        self.segment_handler(message_body)

    def extract_next_round_info(self, message_body):
        segments = message_body.split('|')

        self.made_move = False

        connected_players = [f"{nickname}" for nickname in self.nicknames]
        disconnected_players = [f"[Lost Connection] {nickname}" for nickname in self.disconnected_nicknames]
        players_string = ", ".join(connected_players + disconnected_players)

        self.nicknames_label.config(text=f"Players: {players_string}")
        self.cards_in_hand_label.config(text=f"Cards in hand: {self.cards_in_hand}")
        self.hand_value_label.config(text=f"Hand value: {self.hand_value}")
        if not self.nicknames[0] in self.standing_players:
            self.hit_button.grid(row=6, column=0, pady=10)
            self.stand_button.grid(row=7, column=0, pady=10)

        if segments[3] == "1":
            self.buttons_update()

    def end_the_game(self, message_body):
        print(message_body)
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
        self.hand_value_label.config(text="Soon you will be moved back to the main lobby!")
        self.nicknames_label.config(text="")
        self.game_window.after(10000, self.close_window)

    def close_window(self):
        self.game_window.destroy()
        self.chat_window.deiconify()

    def retrieve_state(self, message_body):
        self.game_started = True
        print("AAAAAAAAAAAAAAAA:", message_body)
        self.segment_handler(message_body)

    def stop_the_game(self):
        self.stop_alert()
        self.close_window()

    def stop_alert(self):
        msg = (
            f"Player has left the game\n"
            f"Game over!\n"
        )

        stop_alter_window = tk.Toplevel(self.parent)
        stop_alter_window.title("Game Stopped")

        label = ttk.Label(stop_alter_window, text=msg)
        label.pack(padx=10, pady=10)

    def update_player_state(self, msg):
        msg_parts = msg.split('|')
        player = msg_parts[0]
        player_state = msg_parts[1]

        if player not in self.disconnected_nicknames and player_state == "0":
            self.disconnected_nicknames.append(player)

            connected_players = [f"{nickname}" for nickname in self.nicknames if nickname not in self.disconnected_nicknames]
            disconnected_players = [f"[Lost Connection] {nickname}" for nickname in self.disconnected_nicknames]
            players_string = ", ".join(connected_players + disconnected_players)

            self.nicknames_label.config(text=f"Players: {players_string}")
        elif player in self.disconnected_nicknames and player_state == "1":
            self.disconnected_nicknames.remove(player)

            connected_players = [f"{nickname}" for nickname in self.nicknames if nickname not in self.disconnected_nicknames]
            disconnected_players = [f"[Lost Connection] {nickname}" for nickname in self.disconnected_nicknames]
            players_string = ", ".join(connected_players + disconnected_players)

            self.nicknames_label.config(text=f"Players: {players_string}")

    def kill_app(self):
        self.game_window.destroy()
        self.chat_window.destroy()

