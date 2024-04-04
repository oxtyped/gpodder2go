CREATE TABLE 'device_sync_groups' (
  id INTEGER PRIMARY KEY AUTOINCREMENT,
  sync_status VARCHAR(20) NOT NULL DEFAULT 'pending',
  sync_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
  created_at varchar(255),
  updated_at varchar(255)
);

CREATE TABLE 'device_sync_group_devices' (
  device_sync_group_id INT NOT NULL REFERENCES device_sync_group(id),
  device_id INT NOT NULL REFERENCES devices(id),
  PRIMARY KEY (device_sync_group_id, device_id)
);
