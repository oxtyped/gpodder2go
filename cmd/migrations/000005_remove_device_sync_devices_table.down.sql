CREATE TABLE 'device_sync_group_devices' (
  device_sync_group_id INT NOT NULL REFERENCES device_sync_group(id),
  device_id INT NOT NULL REFERENCES devices(id),
  PRIMARY KEY (device_sync_group_id, device_id)
);
