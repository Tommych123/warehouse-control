-- Schema: items + items_history
BEGIN;

CREATE TABLE IF NOT EXISTS items (
  id          BIGSERIAL PRIMARY KEY,
  sku         TEXT NOT NULL UNIQUE,
  name        TEXT NOT NULL,
  qty         INTEGER NOT NULL DEFAULT 0,
  location    TEXT,
  created_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  updated_at  TIMESTAMPTZ NOT NULL DEFAULT now()
);

-- updated_at autoupdate
CREATE OR REPLACE FUNCTION set_updated_at()
RETURNS trigger AS $$
BEGIN
  NEW.updated_at = now();
  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_items_set_updated_at ON items;
CREATE TRIGGER trg_items_set_updated_at
BEFORE UPDATE ON items
FOR EACH ROW
EXECUTE FUNCTION set_updated_at();

CREATE TABLE IF NOT EXISTS items_history (
  id          BIGSERIAL PRIMARY KEY,
  item_id     BIGINT NOT NULL,
  action      TEXT NOT NULL CHECK (action IN ('insert','update','delete')),
  actor       TEXT,
  actor_role  TEXT,
  changed_at  TIMESTAMPTZ NOT NULL DEFAULT now(),
  old_data    JSONB,
  new_data    JSONB
);

CREATE INDEX IF NOT EXISTS idx_items_history_item_time
  ON items_history (item_id, changed_at DESC);

CREATE INDEX IF NOT EXISTS idx_items_history_actor
  ON items_history (actor);

CREATE INDEX IF NOT EXISTS idx_items_history_action
  ON items_history (action);

CREATE INDEX IF NOT EXISTS idx_items_history_changed_at
  ON items_history (changed_at);

-- Audit trigger
CREATE OR REPLACE FUNCTION audit_items()
RETURNS trigger AS $$
DECLARE
  v_actor TEXT := NULL;
  v_role  TEXT := NULL;
BEGIN
  v_actor := current_setting('app.user', true);
  v_role  := current_setting('app.role', true);

  IF (TG_OP = 'INSERT') THEN
    INSERT INTO items_history(item_id, action, actor, actor_role, old_data, new_data)
    VALUES (NEW.id, 'insert', v_actor, v_role, NULL, to_jsonb(NEW));
    RETURN NEW;
  ELSIF (TG_OP = 'UPDATE') THEN
    INSERT INTO items_history(item_id, action, actor, actor_role, old_data, new_data)
    VALUES (NEW.id, 'update', v_actor, v_role, to_jsonb(OLD), to_jsonb(NEW));
    RETURN NEW;
  ELSIF (TG_OP = 'DELETE') THEN
    INSERT INTO items_history(item_id, action, actor, actor_role, old_data, new_data)
    VALUES (OLD.id, 'delete', v_actor, v_role, to_jsonb(OLD), NULL);
    RETURN OLD;
  END IF;

  RETURN NULL;
END;
$$ LANGUAGE plpgsql;

DROP TRIGGER IF EXISTS trg_items_audit ON items;
CREATE TRIGGER trg_items_audit
AFTER INSERT OR UPDATE OR DELETE ON items
FOR EACH ROW
EXECUTE FUNCTION audit_items();

COMMIT;
