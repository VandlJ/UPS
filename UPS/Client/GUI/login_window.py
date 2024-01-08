import socket
import threading
from tkinter import ttk
import tkinter as tk
import utils.validation as valid
import utils.client_to_server_message as message_handler
from constants import message_constants
from GUI.lobby_window import LobbyWindow
from GUI.game_window import GameWindow
import time


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

        self.timeout_duration = 10
        self.last_message_time = None
        self.timer_thread = None
        self.timer_stop_event = None
        self.is_server_available = False
        self.player_nickname = None
        self.lock = threading.Lock()

    def update_children_state(self):
        if self.game_window_initializer is not None:
            print("game window update")
            # self.game_window_initializer.update_connection_status(self.is_server_available)
        if self.lobby_window_initializer is not None:
            print("lobby window update")
            # self.lobby_window_initializer.update_connection_status(self.is_server_available)

    def disconnect_from_server(self):
        valid.disconnected_alert(self.root)
        self.lobby_window_initializer.chat_window.destroy()
        if self.game_window_initializer is not None:
            self.game_window_initializer.game_window.destroy()
        try:
            if self.server:
                self.server.close()
        except Exception as e:
            print(f"Error closing the socket: {e}")
        self.server = None
        self.root.deiconify()
        self.timer_stop_event.set()

    def check_timeout(self):
        while not self.timer_stop_event.is_set():
            with self.lock:
                current_time = time.time()
                elapsed_time = current_time - self.last_message_time
                if elapsed_time >= self.timeout_duration:
                    if self.is_server_available:  # change of the state - connected -> not connected
                        valid.connection_lost_alert(self.root)
                    self.is_server_available = False
                else:
                    self.is_server_available = True
                if elapsed_time >= 30:
                    self.disconnect_from_server()
                    return

                self.update_children_state()

            time.sleep(2)

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
            self.response_thread.daemon = True
            self.response_thread.start()

            self.open_chat_window(self.server)

            self.server.sendall((message + "\n").encode())
            self.last_message_time = time.time()

            self.timer_stop_event = threading.Event()
            self.timer_thread = threading.Thread(target=self.check_timeout)
            self.timer_thread.daemon = True
            self.timer_thread.start()

            self.root.withdraw()

        except Exception as e:
            print(f"Error connecting to server: {e}")

    def handle_server_response(self):
        while True:
            try:
                data = self.server.recv(1024)
                self.last_message_time = time.time()
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
        message_type = message[len(message_constants.PASS) + message_constants.FORMAT_LEN:len(
            message_constants.PASS) + message_constants.FORMAT_LEN + message_constants.CMD_LEN]
        message_body = message[len(
            message_constants.PASS) + message_constants.FORMAT_LEN + message_constants.CMD_LEN:]
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
        elif message_type == message_constants.PING:
            self.send_pong()
        elif message_type == message_constants.RETRIEVING_STATE:
            self.game_window_initializer.retrieve_state(message_body)
        elif message_type == message_constants.GAME_STOP:
            self.game_window_initializer.stop_the_game()
        elif message_type == message_constants.PLAYER_STATE:
            self.game_window_initializer.update_player_state(message_body)
        elif message_type == message_constants.OCCUPIED_NICK:
            valid.occupied_nick_alert(self.root)
        elif message_type == message_constants.KILL:
            self.game_window_initializer.kill_app()
        elif message_type == message_constants.KILL2:
            self.lobby_window_initializer.kill_app2()

    def send_pong(self):
        msg = message_handler.create_pong_msg()
        self.server.sendall((msg + "\n").encode())

    def update_game_list(self, games):
        self.lobby_window_initializer.update_game_list(games)
