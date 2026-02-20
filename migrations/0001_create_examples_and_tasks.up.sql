-- ===================================
-- Create function for auto-updated_at
-- ===================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ language 'plpgsql';

-- ===================================
-- Table: users
-- ===================================
CREATE TABLE users (
    id VARCHAR(64) PRIMARY KEY,
    username VARCHAR(50) NOT NULL UNIQUE,
    email VARCHAR(100) NOT NULL UNIQUE,
    password_hash TEXT NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'user', -- 'user', 'admin'
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_users_email ON users(email);
CREATE INDEX idx_users_username ON users(username);

-- ===================================
-- Table: examples
-- ===================================
CREATE TABLE examples (
    id VARCHAR(64) PRIMARY KEY,
    expression TEXT NOT NULL,
    response VARCHAR(64) NOT NULL, -- final result variable
    calculated BOOLEAN NOT NULL DEFAULT FALSE,
    result DOUBLE PRECISION,
    error TEXT, -- e.g., "division by zero"
    user_id VARCHAR(64) NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_examples_user_id ON examples(user_id);
CREATE INDEX idx_examples_calculated ON examples(calculated);

-- ===================================
-- Table: tasks (intermediate steps for report)
-- ===================================
CREATE TABLE tasks (
    id VARCHAR(64) PRIMARY KEY,
    example_id VARCHAR(64) NOT NULL REFERENCES examples(id) ON DELETE CASCADE,
    value1 FLOAT8 NOT NULL,
    value2 FLOAT8 NOT NULL,
    result FLOAT8 NOT NULL,
    sign VARCHAR(2) NOT NULL,
    variable VARCHAR(64) NOT NULL,
    "order" INTEGER NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX idx_tasks_example_id ON tasks(example_id);
CREATE INDEX idx_tasks_variable ON tasks(variable);

-- ===================================
-- Triggers for auto-updated_at
-- ===================================
CREATE TRIGGER update_examples_updated_at
    BEFORE UPDATE ON examples
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_users_updated_at
    BEFORE UPDATE ON users
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();