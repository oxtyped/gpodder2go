# Structure CLI

- gpodder2go serve
- gpodder2go accounts create
- gpodder2go accounts delete

### gpodder2go serve

#### NAME
  gpodder2go serve - starts a gpodder2 server

#### CLI USAGE

```
gpodder2go serve ---database=DB_URL --addr=IP:PORT
```

#### FLAGS

> `--database`=`DB_URL`
>> The database to connect to in a db_uri scheme format

> `--addr`=`IP:PORT`
>> The Addr that the server will bind to

#### EXAMPLES

```
$ gpodder2go serve --database=sqlite3://g2g.db --addr=0.0.0.0:3005
```

### gpodder2go accounts create

#### NAME
  gpodder2go accounts create - creates a user account to connect to gpodder2go

#### CLI USAGE

```
gpodder2go accounts create [NAME] --password=[PASSWORD]
```

#### POSITIONAL ARGUMENTS

> `NAME`
>> Username to create

#### FLAGS

> `--password=`PASSWORD
>> Password for user

#### FLAGS

#### EXAMPLES

```
$ gpodder2go accounts create user1 --password=pass1
```
