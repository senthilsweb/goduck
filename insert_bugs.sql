-- Insert synthetic data into the bugs table
INSERT INTO bugs (id, title, description, status, created_at, modified_at, update_count)
VALUES
    (1, 'Sample Bug Title', 'This is a sample bug description.', 'Closed', '2023-08-29 21:03:23.545206', '2023-08-29 21:18:31.477748', 2),
    (2, 'Another Bug', 'A bug with another description.', 'Open', '2023-08-30 09:00:00', '2023-08-30 09:30:00', 1);
