import tkinter as tk
from tkinter import ttk
from utils.client_to_server_message import create_start_game_message
from utils.client_to_server_message import create_turn_message


# Class responsible for managing the gameplay screen
class GamePlayScreen:
    def __init__(self, master, server, game_list_screen):
        # Initialize GamePlayScreen with necessary attributes
        self.master = master
        self.server = server

        self.game_list_screen = game_list_screen
        self.game_play_screen = None

        self.stand_button = None
        self.hit_button = None
        self.hand_value_label = None
        self.cards_in_hand_label = None
        self.nicknames_label = None
        self.current_players_label = None
        self.start_button = None
        self.status_label = None

        self.start_button_mounted = False
        self._can_be_started = False
        self.game_started = False
        self.game_gui_mounted = False
        self.made_move = False
        self.game_ended = False

        self._current_players = 0
        self._max_players = 0
        self.nicknames = []
        self.disconnected_nicknames = []
        self.standing_players = []
        self.cards_in_hand = ""
        self.hand_value = ""

        self.winners = []

    # Properties to manage current players, max players, and game readiness
    @property
    def current_players(self):
        return self._current_players

    @current_players.setter
    def current_players(self, value):
        if self._current_players != value:
            self._current_players = value
            self.gui_update()

    @property
    def max_players(self):
        return self._max_players

    @max_players.setter
    def max_players(self, value):
        if self._max_players != value:
            self._max_players = value
            self.gui_update()

    @property
    def game_ready(self):
        return self._can_be_started

    @game_ready.setter
    def game_ready(self, value):
        if self._can_be_started != value:
            self._can_be_started = value
            self.gui_update()

    def game_init(self):
        # Initialize the game GUI when the game starts
        if not self.game_gui_mounted:
            self.status_label.config(text="Game started")

            self.nicknames_label = ttk.Label(self.game_play_screen, text="Players: ")
            self.nicknames_label.grid(row=3, column=0, pady=5)

            self.cards_in_hand_label = ttk.Label(self.game_play_screen, text="Cards: ")
            self.cards_in_hand_label.grid(row=4, column=0, pady=5)

            self.hand_value_label = ttk.Label(self.game_play_screen, text="Hand value: ")
            self.hand_value_label.grid(row=5, column=0, pady=5)

            self.start_button.grid_forget()

            self.update_player_info()

            self.game_gui_mounted = True

        players_info = f"{self._current_players}/{self._max_players}"
        self.current_players_label.config(text=f"Current players: {players_info}")

        self.update_player_info()

        self.hit_button = tk.Button(self.game_play_screen, text="Hit", command=lambda: self.send_move("HIT"))
        self.stand_button = tk.Button(self.game_play_screen, text="Stand", command=lambda: self.send_move("STAND"))

        self.hit_button.grid(row=6, column=0, pady=10)
        self.stand_button.grid(row=7, column=0, pady=10)

    def gui_update(self):
        # Update the GUI based on game status and actions performed
        if self.game_ready and not self.game_started and not self.made_move:
            if not self.start_button_mounted:
                self.start_button.grid(row=1, column=0, pady=10)
                self.start_button_mounted = True
            self.current_players_update()
        elif self.game_started and not self.made_move:
            self.game_init()
        elif self.made_move and not self.game_started:
            self.buttons_update()
        elif self.game_ended:
            self.end_screen()
        else:
            self.current_players_update()

    def open_game_window(self):
        # Open the game window and configure its layout
        self.game_play_screen = tk.Toplevel(self.master)
        self.game_play_screen.title("Blackjack Game Window")
        self.game_play_screen.protocol("WM_DELETE_WINDOW", self.destroy_parent)

        self.status_label = tk.Label(self.game_play_screen, text="Waiting for players")
        self.status_label.grid(row=0, column=0, pady=10)

        self.start_button = tk.Button(self.game_play_screen, text="Start the Game", command=self.start_game)

        self.current_players_label = tk.Label(self.game_play_screen, text="Current players: ")
        self.current_players_label.grid(row=2, column=0, pady=5)

        self.gui_update()

    def extract_next_round_info(self, message_body):
        # Extract and handle information for the next round of the game
        segments = message_body.split('|')

        self.made_move = False

        self.update_player_info()

        if not self.nicknames[0] in self.standing_players:
            self.hit_button.grid(row=6, column=0, pady=10)
            self.stand_button.grid(row=7, column=0, pady=10)

        if segments[3] == "1":
            self.buttons_update()

    def end_screen(self):
        # Handle the end of the game, display winners, and finalize the game
        if len(self.winners) > 1:
            self.cards_in_hand_label.config(text="Winners: " + ";".join(self.winners))
        else:
            self.cards_in_hand_label.config(text="Winner: " + self.winners[0])
        self.status_label.config(text="Game has ended")
        self.hand_value_label.config(text="Wait for moving to game list screen")
        self.nicknames_label.config(text="")
        self.game_play_screen.after(10000, self.close_window)

    def update_player_state(self, msg):
        # Update the player's connection state (connected or disconnected)
        msg_parts = msg.split('|')
        player = msg_parts[0]
        player_state = msg_parts[1]

        if player not in self.disconnected_nicknames and player_state == "0":
            self.disconnected_nicknames.append(player)

            connected_players = [f"{nickname}" for nickname in self.nicknames if nickname
                                 not in self.disconnected_nicknames]
            disconnected_players = [f"Disconnected - {nickname}" for nickname in self.disconnected_nicknames]
            players_string = ", ".join(connected_players + disconnected_players)

            self.nicknames_label.config(text=f"Players: {players_string}")
        elif player in self.disconnected_nicknames and player_state == "1":
            self.disconnected_nicknames.remove(player)

            connected_players = [f"{nickname}" for nickname in self.nicknames if nickname
                                 not in self.disconnected_nicknames]
            disconnected_players = [f"Disconnected -  {nickname}" for nickname in self.disconnected_nicknames]
            players_string = ", ".join(connected_players + disconnected_players)

            self.nicknames_label.config(text=f"Players: {players_string}")

    def segment_handler(self, message_body):
        # Handle various segments of messages received from the server
        segments = message_body.split('|')
        if len(segments) == 3:
            self.segment_splitter(segments)
            self.gui_update()
        elif len(segments) == 4:
            self.segment_splitter(segments)
            if segments[3] == "1":
                self.gui_update()
                self.buttons_update()
            else:
                self.gui_update()
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

    # Methods to handle player actions (e.g., sending moves, starting the game)
    def send_move(self, turn):
        if turn == "STAND":
            self.standing_players.append(self.nicknames[0])

        message = create_turn_message(turn)
        self.server.sendall((message + "\n").encode())
        self.made_move = True
        self.game_started = False
        self.gui_update()

    def start_game(self):
        message = create_start_game_message()
        self.server.sendall((message + "\n").encode())

    # Methods to update GUI elements and manage window destruction
    def buttons_update(self):
        self.hit_button.grid_forget()
        self.stand_button.grid_forget()

    def destroy_parent(self):
        self.master.destroy()

    def current_players_update(self):
        players_info = f"{self._current_players}/{self._max_players}"
        self.current_players_label.config(text=f"Current players: {players_info}")

    def extract_turn_info(self, message_body):
        self.game_started = False
        self.segment_handler(message_body)

    def extract_init_game_info(self, message_body):
        self.game_started = True
        self.segment_handler(message_body)

    def end_the_game(self, message_body):
        print(message_body)
        nicknames = message_body.split(';')
        self.winners = nicknames
        self.game_ended = True
        self.gui_update()

    def close_window(self):
        self.game_play_screen.destroy()
        self.game_list_screen.deiconify()

    def retrieve_state(self, message_body):
        # Retrieve the game state to continue the game
        self.game_started = True
        self.segment_handler(message_body)

    # Methods to handle stopping or killing the game
    def stop_the_game(self):
        self.stop_alert()
        self.close_window()

    def kill_app(self):
        self.game_play_screen.destroy()
        self.game_list_screen.deiconify()

    def stop_alert(self):
        msg = (
            f"Someone disconnected\n"
            f"Game has ended\n"
        )

        stop_alter_window = tk.Toplevel(self.master)
        stop_alter_window.title("Game Stopped")

        label = ttk.Label(stop_alter_window, text=msg)
        label.pack(padx=10, pady=10)

    def update_player_info(self):
        # Update the displayed information about players in the game
        connected_players = [f"{nickname}" for nickname in self.nicknames]
        disconnected_players = [f"Disconnected - {nickname}" for nickname in self.disconnected_nicknames]
        players_string = ", ".join(connected_players + disconnected_players)

        self.nicknames_label.config(text=f"Players: {players_string}")
        self.cards_in_hand_label.config(text=f"Cards in hand: {self.cards_in_hand}")
        self.hand_value_label.config(text=f"Hand value: {self.hand_value}")
