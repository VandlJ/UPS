from constants import msg_const


def get_ping_pong_interval(message):
    # Extracts the ping-pong interval from the message received
    parts = message.split('|')

    if len(parts) == 2:
        try:
            interval = parts[1]
            return interval
        except ValueError:
            return None
    else:
        return None


def create_pong_msg():
    # Creates a message to respond to a ping from the server
    msg = f"{msg_const.PASS}000{msg_const.PONG}"
    return msg


def create_nick_msg(nick):
    # Creates a message to set the nickname for the game
    nick_len = str(len(nick)).zfill(3)
    msg = f"{msg_const.PASS}{nick_len}{msg_const.NICK}" f"{nick}"
    return msg


def msg_valid_checker(message):
    # Checks the validity of the received message based on predefined constants
    if len(message) < (len(msg_const.PASS) + msg_const.CMD_LEN +
                       msg_const.FORMAT_LEN):
        return False

    password = message[:len(msg_const.PASS)]
    if password != msg_const.PASS:
        print(f"Password {password}, constant {msg_const.PASS}")
        return False

    length_string = message[len(msg_const.PASS):len(msg_const.PASS) + msg_const.FORMAT_LEN]
    try:
        length = int(length_string)
    except ValueError:
        return False

    if __name__ == '__main__':
        if (length != len(message) - len(msg_const.PASS) -
                msg_const.FORMAT_LEN - msg_const.CMD_LEN):
            print(f"Length from message: {length}, calculated length: "
                  f"{len(message) - len(msg_const.PASS) - msg_const.FORMAT_LEN - msg_const.CMD_LEN}")
            return False

    return True


def games_info_extractor(message):
    # Extracts game information from the message received from the server
    game_strings = message.split(';')
    games = []

    for game_string in game_strings:
        game_components = game_string.split('|')
        segment_split(game_components, games)

    return games


def players_extractor(message):
    # Extracts the number of current players and the maximum players allowed in a game
    parts = message.split('|')

    if len(parts) == 3:
        try:
            current_players = int(parts[1])
            max_players = int(parts[2])
            return current_players, max_players
        except ValueError:
            return None, None
    else:
        return None, None


def join_msg_creator(game_name):
    # Creates a message to join a particular game
    len_game_name = str(len(game_name)).zfill(3)
    formatted_message = f"{msg_const.PASS}{len_game_name}{msg_const.JOIN}{game_name}"
    print("Formatted message: ", formatted_message)
    return formatted_message


def game_start_msg_creator():
    # Creates a message to start the game
    msg = f"{msg_const.PASS}000{msg_const.PLAY}"
    return msg


def turn_msg_creator(turn):
    # Creates a message to indicate a turn action in the game
    nick_len = str(len(turn)).zfill(3)
    msg = f"{msg_const.PASS}{nick_len}{msg_const.TURN}{turn}"
    return msg


def segment_split(game_components, games):
    # Splits game components and structures game information
    if len(game_components) == 4:
        game_name, max_players, current_players, game_status = game_components
        game_info = {
            'game_name': game_name,
            'max_players': int(max_players),
            'current_players': int(current_players),
            'game_status': int(game_status),
        }
        games.append(game_info)


def join_success(msg):
    # Checks if a join to a game is successful
    return msg == "1"


def game_check(msg):
    # Checks the status of the game
    return msg[0] == "1"
