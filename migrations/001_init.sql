-- USERS
CREATE TABLE users (
    id SERIAL PRIMARY KEY,
    username TEXT UNIQUE NOT NULL,
    role TEXT NOT NULL CHECK (role IN ('admin', 'manager', 'viewer'))
);

INSERT INTO users(username, role) VALUES
('admin', 'admin'),
('manager', 'manager'),
('viewer', 'viewer');

-- ITEMS
CREATE TABLE items (
    id SERIAL PRIMARY KEY,
    name TEXT NOT NULL,
    quantity INT NOT NULL,
    description TEXT,
    updated_at TIMESTAMP DEFAULT now()
);

-- HISTORY
CREATE TABLE item_history (
    id SERIAL PRIMARY KEY,
    item_id INT,
    action TEXT NOT NULL,
    old_data JSONB,
    new_data JSONB,
    changed_by TEXT,
    changed_at TIMESTAMP DEFAULT now()
);

-- TRIGGER FUNCTION (АНТИПАТТЕРН)
CREATE OR REPLACE FUNCTION log_item_changes()
RETURNS TRIGGER AS $$
BEGIN
    IF TG_OP = 'INSERT' THEN
        INSERT INTO item_history(item_id, action, new_data, changed_by)
        VALUES (
            NEW.id,
            'INSERT',
            to_jsonb(NEW),
            current_setting('app.user', true)
        );

    ELSIF TG_OP = 'UPDATE' THEN
        INSERT INTO item_history(item_id, action, old_data, new_data, changed_by)
        VALUES (
            NEW.id,
            'UPDATE',
            to_jsonb(OLD),
            to_jsonb(NEW),
            current_setting('app.user', true)
        );

    ELSIF TG_OP = 'DELETE' THEN
        INSERT INTO item_history(item_id, action, old_data, changed_by)
        VALUES (
            OLD.id,
            'DELETE',
            to_jsonb(OLD),
            current_setting('app.user', true)
        );
    END IF;

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- TRIGGER
CREATE TRIGGER trg_items_history
AFTER INSERT OR UPDATE OR DELETE ON items
FOR EACH ROW EXECUTE FUNCTION log_item_changes();
