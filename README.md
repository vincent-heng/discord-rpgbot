# Introduction

*RPGBot* is a Discord Bot for textual multiplayer RPG. A Game Master is required.

# Installation

- Clone this project
```
git clone https://github.com/vincent-heng/discord-rpgbot
```

- Run it with Docker Compose. It will also install a persistent postgres db in a container.
```
docker-compose up
```

You can run psql commands on the psql shell
```
docker exec -it rpgbot_db_1 psql -U postgres
```


# Project Structure
```
/
        /scripts : DB scripts, like database initialization
        rpgbot.go : main file, with bot behaviour
        service.go : bot behaviour functions
        utils.go : utilities functions
        config.json : configuration parameters. Copy it from config-sample.json
```
