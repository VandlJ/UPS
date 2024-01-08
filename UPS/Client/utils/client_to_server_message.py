from constants import msg_const


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
    formatted_msg = f"{msg_const.PASS}000{msg_const.PONG}"
    return formatted_msg


def create_nick_msg(nickname):
    nickname_len = str(len(nickname)).zfill(3)
    formatted_message = (f"{msg_const.PASS}{nickname_len}{msg_const.NICK}"
                         f"{nickname}")
    return formatted_message


def msg_valid_checker(message):
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


def join_msg_creator(game_name):
    len_game_name = str(len(game_name)).zfill(3)
    formatted_message = f"{msg_const.PASS}{len_game_name}{msg_const.JOIN_TYPE}{game_name}"
    print("Formatted message: ", formatted_message)
    return formatted_message


def create_start_game_message():
    formatted_message = f"{msg_const.PASS}000{msg_const.START_THE_GAME}"
    return formatted_message


def create_turn_message(turn):
    len_nick = str(len(turn)).zfill(3)
    formatted_message = f"{msg_const.PASS}{len_nick}{msg_const.TURN}{turn}"
    return formatted_message
