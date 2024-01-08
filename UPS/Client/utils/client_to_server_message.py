from constants import message_constants


def get_ping_pong_interval(message):
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
    formatted_msg = f"{message_constants.PASS}000{message_constants.PONG}"
    return formatted_msg


def create_nickname_message(nickname):
    nickname_len = str(len(nickname)).zfill(3)
    formatted_message = (f"{message_constants.PASS}{nickname_len}{message_constants.NICK}"
                         f"{nickname}")
    return formatted_message


def is_message_valid(message):
    if len(message) < (len(message_constants.PASS) + message_constants.CMD_LEN +
                       message_constants.FORMAT_LEN):
        return False

    password = message[:len(message_constants.PASS)]
    if password != message_constants.PASS:
        print(f"Password {password}, constant {message_constants.PASS}")
        return False

    length_string = message[len(message_constants.PASS):len(message_constants.PASS) + message_constants.FORMAT_LEN]
    try:
        length = int(length_string)
    except ValueError:
        return False

    if __name__ == '__main__':
        if (length != len(message) - len(message_constants.PASS) -
                message_constants.FORMAT_LEN - message_constants.CMD_LEN):
            print(f"Length from message: {length}, calculated length: "
                  f"{len(message) - len(message_constants.PASS) - message_constants.FORMAT_LEN - message_constants.CMD_LEN}")
            return False

    return True


def extract_games_info(message):
    game_strings = message.split(';')
    games = []

    for game_string in game_strings:
        game_components = game_string.split('|')
        if len(game_components) == 4:
            game_name, max_players, current_players, game_status = game_components
            game_info = {
                'game_name': game_name,
                'max_players': int(max_players),
                'current_players': int(current_players),
                'game_status': int(game_status),
            }

            games.append(game_info)

    return games


def joined_game_successfully(message):
    return message == "1"


def can_game_begin(message):
    print("Message: ", message)
    return message[0] == "1"


def extract_players(message):
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


def create_game_joining_message(game_name):
    len_game_name = str(len(game_name)).zfill(3)
    formatted_message = f"{message_constants.PASS}{len_game_name}{message_constants.JOIN_TYPE}{game_name}"
    print("Formatted message: ", formatted_message)
    return formatted_message


def create_start_game_message():
    formatted_message = f"{message_constants.PASS}000{message_constants.START_THE_GAME}"
    return formatted_message


def create_turn_message(turn):
    len_nick = str(len(turn)).zfill(3)
    formatted_message = f"{message_constants.PASS}{len_nick}{message_constants.TURN}{turn}"
    return formatted_message
