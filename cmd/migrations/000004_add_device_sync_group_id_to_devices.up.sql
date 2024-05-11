ALTER TABLE devices
ADD COLUMN device_sync_group_id INTEGER
REFERENCES device_sync_groups(id);
