CREATE TABLE 'users' (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  username varchar (50) UNIQUE NOT NULL,
  password varchar (50) NOT NULL,
  email varchar (255) UNIQUE NOT NULL,
  name varchar(50) NOT NULL,
  created_at varchar(255),
  updated_at varchar(255)
);

CREATE TABLE 'devices' (
id INTEGER PRIMARY KEY AUTOINCREMENT,
user_id INT NOT NULL,
name varchar(255) NOT NULL,
type varchar(255) NOT NULL,
caption varchar(255),
created_at varchar(255),
updated_at varchar(255),
FOREIGN KEY (user_id) REFERENCES users(id)
);

CREATE TABLE 'subscriptions' (
id INTEGER PRIMARY KEY AUTOINCREMENT,
device_id INT NOT NULL,
user_id INT NOT NULL,
podcast varchar(255) NOT NULL,
action varchar(100) NOT NULL,
timestamp varchar(255),
created_at varchar(255),
updated_at varchar(255),
FOREIGN KEY (device_id) REFERENCES devices (id),
FOREIGN KEY (user_id) REFERENCES users (id)
);

CREATE TABLE 'episode_actions' (
id INTEGER PRIMARY KEY AUTOINCREMENT,
device_id INT NOT NULL,
podcast varchar(255) NOT NULL,
episode varchar(255) NOT NULL,
action varchar(100) NOT NULL,
position int,
started int,
total int,
created_at varchar(255),
updated_at varchar(255),
timestamp varchar(255),
FOREIGN KEY (device_id) REFERENCES devices(id)
);
