#!/usr/bin/env sh

DATABASE_PATH=/app/data/db/LOCAL_BINDING.sqlite3

# command add
# adds a user to the sqlite database
# usage: add <username> <password> <email>

function add() {
    if [ -z "$1" ]; then
        echo "Error: username is required"
        exit 1
    fi
    if [ -z "$2" ]; then
        echo "Error: password is required"
        exit 1
    fi
    if [ -z "$3" ]; then
        echo "Error: email is required"
        exit 1
    fi

    if [ ! -f "$DATABASE_PATH" ]; then
        echo "Error: database does not exist"
        exit 1
    fi

    PASSWORD=$(htpasswd -bnBC 10 "" "$2" | tr -d ':\n')

    echo "INSERT INTO user (username, password, email, verify_code) VALUES ('$1', '$PASSWORD', '$3', NULL);" | sqlite3 "$DATABASE_PATH"
}

# command del
# deletes a user from the sqlite database
# usage: del <username>

function del() {
    if [ -z "$1" ]; then
        echo "Error: email is required"
        exit 1
    fi

    if [ ! -f "$DATABASE_PATH" ]; then
        echo "Error: database does not exist"
        exit 1
    fi

    echo "DELETE FROM user WHERE email = '$1';" | sqlite3 "$DATABASE_PATH"
}

# command list
# lists all users in the sqlite database
# usage: list

function list() {
    if [ ! -f "$DATABASE_PATH" ]; then
        echo "Error: database does not exist"
        exit 1
    fi

    echo "SELECT * FROM user;" | sqlite3 "$DATABASE_PATH"
}

# command activate
# activates a user in the sqlite database
# usage: activate <username>

function activate() {
    if [ -z "$1" ]; then
        echo "Error: email is required"
        exit 1
    fi

    if [ ! -f "$DATABASE_PATH" ]; then
        echo "Error: database does not exist"
        exit 1
    fi

    echo "UPDATE user SET verify_code = NULL WHERE email = '$1';" | sqlite3 "$DATABASE_PATH"
}

# command help
# prints this help message
# usage: help

function help() {
    echo "Usage: $0 <command> [args]"
    echo "Commands:"
    echo "  add <username> <password> <email> - adds a user to the sqlite database"
    echo "  del <email> - deletes a user from the sqlite database"
    echo "  activate <email> - activates a user in the sqlite database"
    echo "  list - lists all users in the sqlite database"
    echo "  help - prints this help message"
}


# switch case commands

case "$1" in
    add)
        add "$2" "$3" "$4"
        ;;
    del)
        del "$2"
        ;;
    list)
        list
        ;;
    help)
        help
        ;;
    *)
        help
        ;;
esac