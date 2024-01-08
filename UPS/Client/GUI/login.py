import socket
import threading
import time
from tkinter import ttk

from constants import msg_const
from GUI.game_list import GameListScreen
from GUI.game_play import GamePlayScreen

import utils.validation as valid
import utils.client_to_server_message as msg_handler


# Class responsible for managing the login screen
class LoginScreen:
    def __init__(self, master):
        # Initialize the login screen with various attributes
        self.master = master

        self.server = None
        self.communication_thread = None
        self.game_list_listbox = None
        self.game_list_screen_init = None
        self.game_play_screen_init = None

        self.server_state = False

        self.last_message_time = None
        self.timer_thread = None
        self.timer_stop_event = None
        self.thread_locker = threading.Lock()

        self.buffer = b''

        self.master.title("Blackjack")

        self.server_ip_label = ttk.Label(master, text="Server IP:")
        self.server_ip_label.grid(row=0, column=0, padx=5, pady=5, sticky="w")
        self.server_ip_entry = ttk.Entry(master)
        self.server_ip_entry.grid(row=0, column=1, padx=5, pady=5)

        self.server_port_label = ttk.Label(master, text="Server port:")
        self.server_port_label.grid(row=1, column=0, padx=5, pady=5, sticky="w")
        self.server_port_entry = ttk.Entry(master)
        self.server_port_entry.grid(row=1, column=1, padx=5, pady=5)

        self.nickname_label = ttk.Label(master, text="Nickname:")
        self.nickname_label.grid(row=2, column=0, padx=5, pady=5, sticky="w")
        self.nick_entry = ttk.Entry(master)
        self.nick_entry.grid(row=2, column=1, padx=5, pady=5)

        self.connect_button = ttk.Button(master, text="Connect", command=self.server_connector)
        self.connect_button.grid(row=3, column=0, padx=5, pady=10)

    def msg_tree(self, msg):
        # Handle different types of messages received from the server
        # and perform actions based on the message content

        self.last_message_time = time.time()
        if not self.server_state:
            if self.game_play_screen_init is not None:
                self.game_play_screen_init.resend_state()
        self.server_state = True

        cmd = msg[len(msg_const.PASS) + msg_const.FORMAT_LEN:len(
            msg_const.PASS) + msg_const.FORMAT_LEN + msg_const.CMD_LEN]
        msg_cnt = msg[len(
            msg_const.PASS) + msg_const.FORMAT_LEN + msg_const.CMD_LEN:]
        if cmd == msg_const.GAME_INFO:
            games = msg_handler.extract_games_info(msg_cnt)
            self.update_game_list(games)
        elif cmd == msg_const.LOBBY_JOIN_RESPONSE:
            success = msg_handler.joined_game_successfully(msg_cnt)
            if success:
                self.game_list_screen_init.withdraw_game_list_screen()
                self.game_play_screen_init = GamePlayScreen(self.master, self.server,
                                                            self.game_list_screen_init.game_list_screen)
                self.game_play_screen_init.open_game_window()
        elif cmd == msg_const.CAN_GAME_START:
            can_be_started = msg_handler.can_game_begin(msg_cnt)
            current_players, max_players = msg_handler.extract_players(msg_cnt)

            self.game_play_screen_init.current_players = current_players
            self.game_play_screen_init.max_players = max_players
            self.game_play_screen_init.can_be_started = can_be_started

        elif cmd == msg_const.GAME_STARTED_INIT:
            self.game_play_screen_init.extract_init_game_info(msg_cnt)
        elif cmd == msg_const.TURN:
            self.game_play_screen_init.extract_turn_info(msg_cnt)
        elif cmd == msg_const.NEXT_ROUND:
            self.game_play_screen_init.extract_next_round_info(msg_cnt)
        elif cmd == msg_const.GAME_ENDING:
            self.game_play_screen_init.end_the_game(msg_cnt)
        elif cmd == msg_const.PING:
            self.send_pong()
        elif cmd == msg_const.RETRIEVING_STATE:
            self.game_play_screen_init.retrieve_state(msg_cnt)
        elif cmd == msg_const.GAME_STOP:
            self.game_play_screen_init.stop_the_game()
        elif cmd == msg_const.PLAYER_STATE:
            self.game_play_screen_init.update_player_state(msg_cnt)
        elif cmd == msg_const.OCCUPIED_NICK:
            valid.occupied_nick_alert(self.master)
        elif cmd == msg_const.KILL:
            self.game_play_screen_init.kill_app()
        elif cmd == msg_const.KILL2:
            self.game_list_screen_init.kill_app2()

    def server_disconnector(self):
        # Disconnect the client from the server and perform necessary cleanup
        valid.disconnected_alert(self.master)
        self.game_list_screen_init.game_list_screen.destroy()
        if self.game_play_screen_init is not None:
            self.game_play_screen_init.game_play_screen.destroy()
        try:
            if self.server:
                self.server.close()
        except Exception as e:
            print(f"Killing: {e}")
        self.server = None
        self.master.deiconify()
        self.timer_stop_event.set()

    def server_connector(self):
        # Connect the client to the server using provided server IP, port, and nickname
        server_ip = self.server_ip_entry.get()
        server_port = int(self.server_port_entry.get())
        nick = self.nick_entry.get()

        if not (valid.validate_server_ip(server_ip) and valid.validate_server_port(server_port)
                and valid.validate_nick(nick)):
            return

        msg = msg_handler.create_nick_msg(self.nick_entry.get())

        try:
            self.server = socket.socket(socket.AF_INET, socket.SOCK_STREAM)
            self.server.connect((server_ip, server_port))

            self.communication_thread = threading.Thread(target=self.server_msg_checker)
            self.communication_thread.daemon = True
            self.communication_thread.start()

            self.game_list_screen_open(self.server)

            self.server.sendall((msg + "\n").encode())
            self.last_message_time = time.time()

            self.timer_stop_event = threading.Event()
            self.timer_thread = threading.Thread(target=self.timeout)
            self.timer_thread.daemon = True
            self.timer_thread.start()

            self.server_state = True

            self.master.withdraw()

        except Exception as e:
            print(f"Error connecting to server: {e}")

    def timeout(self):
        # Monitor the elapsed time since the last server message and handle timeout events
        while not self.timer_stop_event.is_set():
            with self.thread_locker:
                current_time = time.time()
                elapsed_time = current_time - self.last_message_time
                if elapsed_time >= msg_const.TIMEOUT:
                    if self.server_state:
                        valid.connection_lost_alert(self.master)
                    self.server_state = False
                else:
                    self.server_state = True
                if elapsed_time >= msg_const.TIMEOUT_LIMIT:
                    self.server_disconnector()
                    return

            time.sleep(2)

    def server_msg_checker(self):
        # Check for incoming messages from the server, handle message buffering,
        # and process received messages
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
                    self.server_msg_handler(msg)

            except Exception as e:
                print(f"Error reading response from server: {str(e)}")
                break

    def game_list_screen_open(self, server):
        # Open the game list screen with the server connection provided
        self.game_list_screen_init = GameListScreen(self.master, server)
        self.game_list_screen_init.game_list_screen_open()

    def server_msg_handler(self, msg):
        # Handle the received server message by validating it and performing appropriate actions
        msg = msg.replace("\n", "")
        if msg_handler.msg_valid_checker(msg):
            print("Server response:", msg)
            self.msg_tree(msg)
        else:
            print("Message is invalid.")
        pass

    def send_pong(self):
        # Send a 'pong' message to the server as a response to a 'ping' message
        msg = msg_handler.create_pong_msg()
        self.server.sendall((msg + "\n").encode())

    def update_game_list(self, games):
        # Update the game list based on the information received from the server
        self.game_list_screen_init.update_game_list(games)
