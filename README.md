# Key Keeper

An efficient database remote  key management system for database encryption.

Key management is supported for all major DBMS in the form of unified API calls.

Now supported (You need to pre-install the DBMS plugin):

- [x] [MariaDB](https://github.com/RicheyJang/server_key_management)
- [ ] Mysql
- [ ] Postgresql

You can self develop adapter database plugins to support more DBMS.

## install

### prepare

- An intstalled DBMS and key keeper plugin

- CA certificate

- Server Certificate and private key file

- Client Certificate and private key files

### Start

1. Create a new database for key keeper.
2. Download latest [release](https://github.com/RicheyJang/key_keeper/releases) of Key Keeper.
3. Run it first time.
4. Config `config.toml` and restart.
5. Visit https://localhost:7710
6. Create a new instance and key for the database you need to encrypt.

That's all.

## Config

config.toml

```toml
config = "./config.toml"
host = ":7709"  # the host and port of key distribution monitor
web = ":7710"   # the host and port of web UI

[cert]
  ca = "cert/ca.crt"  # the path of CA certificate
  private = "cert/server_rsa_private.pem" # the path of Server private key file
  self = "cert/server.crt" # the path of Server Certificate

[db]
  host = "localhost" # the host of DBMS for key keeper
  name = "kk"        # the name of database for key keeper
  password = "admin" # the user's password of DBMS for key keeper
  port = 5432        # the port of DBMS for key keeper
  type = "postgresql"# the type of DBMS for key keeper, support postgresql and mysql.
  user = "postgres"  # the username of DBMS for key keeper

[log]
  date = 5       # log file retention duration
  dir = "log"    # the dir of log files
  level = "info" # log level

[user]
  maxage = "10h" # user session timeout duration of Web UI
```

