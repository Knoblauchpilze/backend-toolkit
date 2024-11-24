
CREATE TABLE my_table (
  id UUID NOT NULL,
  name TEXT NOT NULL,
  updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
  version INTEGER DEFAULT 0,
  PRIMARY KEY (id),
  UNIQUE (name)
);

CREATE TRIGGER trigger_my_table_updated_at
  BEFORE UPDATE OR INSERT ON my_table
  FOR EACH ROW
  EXECUTE FUNCTION update_updated_at();
