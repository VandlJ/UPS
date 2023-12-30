import socket
import threading
from tkinter import ttk
import utils.validation as valid
import utils.client_to_server_message as message_handler
from constants import message_constants
from GUI.lobby_window import LobbyWindow
from GUI.game_window import GameWindow


class LoginWindow:
    def __init__(self, root):
        self.root = root
        self.root.title("Blackjack Login Window")

        self.server_ip_label = ttk.Label(root, text="Server IP:")
        self.server_ip_label.grid(row=0, column=0, padx=5, pady=5, sticky="w")
        self.server_ip_entry = ttk.Entry(root)
        self.server_ip_entry.grid(row=0, column=1, padx=5, pady=5)

        self.server_port_label = ttk.Label(root, text="Server port:")
        self.server_port_label.grid(row=1, column=0, padx=5, pady=5, sticky="w")
        self.server_port_entry = ttk.Entry(root)
        self.server_port_entry.grid(row=1, column=1, padx=5, pady=5)

        self.nickname_label = ttk.Label(root, text="Nickname:")
        self.nickname_label.grid(row=2, column=0, padx=5, pady=5, sticky="w")
        self.nickname_entry = ttk.Entry(root)
        self.nickname_entry.grid(row=2, column=1, padx=5, pady=5)

        self.connect_button = ttk.Button(root, text="Connect", command=self.connect_to_server)
        self.connect_button.grid(row=3, column=0, padx=5, pady=10)

        self.server = None
        self.response_thread = None
        self.lobby_listbox = None
        self.lobby_window_initializer = None
        self.game_window_initializer = None

        self.buffer = b''

    def connect_to_server(self):
        server_ip = self.server_ip_entry.get()
        server_port = int(self.server_port_entry.get())
        nickname = self.nickname_entry.get()

        if not (valid.validate_server_ip(server_ip) and valid.validate_server_port(server_port)
                and valid.validate_nickname(nickname)):
            return

        message = message_handler.create_nickname_message(self.nickname_entry.get())

        try:
            self.server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.server.connect((server_ip, server_port))

            self.response_thread = threading.Thread(target=self.handle_server_response)
            self.response_thread.start()

            self.open_chat_window(self.server)

            self.server.sendall((message + "\n").encode())

            self.root.withdraw()

        except Exception as e:
            print(f"Error connecting to server: {e}")

    def handle_server_response(self):
        while True:
            try:
                data = self.server.recv(1024)
                if not data:
                    print("Server has disconnected.")
                    break

                self.buffer += data

                messages = self.buffer.split(b'\n')
                self.buffer = messages.pop()

                for msg in messages:
                    msg = msg.decode()
                    self.handle_response_from_server(msg)

            except Exception as e:
                print(f"Error reading response from server: {str(e)}")
                break

    def open_chat_window(self, server):
        self.lobby_window_initializer = LobbyWindow(self.root, server)
        self.lobby_window_initializer.open_chat_window()

    def handle_response_from_server(self, response):
        print("Server response:", response)
        response = response.replace("\n", "")
        if message_handler.is_message_valid(response):
            self.handle_message(response)
        else:
            print("Message is invalid.")
        pass

    def handle_message(self, message):
        message_type = message[len(message_constants.PASSWORD) + message_constants.MESSAGE_LENGTH_FORMAT:len(
            message_constants.PASSWORD) + message_constants.MESSAGE_LENGTH_FORMAT + message_constants.MESSAGE_TYPE_LENGTH]
        message_body = message[len(
            message_constants.PASSWORD) + message_constants.MESSAGE_LENGTH_FORMAT + message_constants.MESSAGE_TYPE_LENGTH:]
        if message_type == message_constants.GAME_INFO:
            games = message_handler.extract_games_info(message_body)
            self.update_game_list(games)
        elif message_type == message_constants.LOBBY_JOIN_RESPONSE:
            success = message_handler.joined_game_successfully(message_body)
            if success:
                self.lobby_window_initializer.close_lobby_window()
                self.game_window_initializer = GameWindow(self.root, self.server,
                                                          self.lobby_window_initializer.chat_window)
                self.game_window_initializer.open_game_window()
        elif message_type == message_constants.CAN_GAME_START:
            can_be_started = message_handler.can_game_begin(message_body)
            current_players, max_players = message_handler.extract_players(message_body)

            self.game_window_initializer.current_players = current_players
            self.game_window_initializer.max_players = max_players
            self.game_window_initializer.can_be_started = can_be_started

        elif message_type == message_constants.GAME_STARTED_INIT:
            self.game_window_initializer.extract_init_game_info(message_body)
        elif message_type == message_constants.TURN:
            self.game_window_initializer.extract_turn_info(message_body)
        elif message_type == message_constants.NEXT_ROUND:
            self.game_window_initializer.extract_next_round_info(message_body)
        elif message_type == message_constants.GAME_ENDING:
            self.game_window_initializer.end_the_game(message_body)

    def update_game_list(self, games):
        self.lobby_window_initializer.update_game_list(games)
