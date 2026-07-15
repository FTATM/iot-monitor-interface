CREATE TABLE IF NOT EXISTS widget (
    widget_id SERIAL,
    widget_type_id INT NOT NULL,
    canvas_id INT NOT NULL,
    layout_data JSONB NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT NULL,
    CONSTRAINT pk_widget PRIMARY KEY (widget_id)
);

CREATE TABLE IF NOT EXISTS widget_type (
    widget_type_id SERIAL,
    widget_type_name TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_widget_type PRIMARY KEY (widget_type_id),
    CONSTRAINT uq_widget_type UNIQUE (widget_type_name)
);

-- preload
INSERT INTO widget_type (widget_type_name) VALUES ('BarChart');
INSERT INTO widget_type (widget_type_name) VALUES ('BulletChart');
INSERT INTO widget_type (widget_type_name) VALUES ('GaugeChart');


CREATE TABLE IF NOT EXISTS canvas (
    canvas_id SERIAL,
    canvas_name TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT NULL,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    CONSTRAINT pk_canvas PRIMARY KEY (canvas_id)
);

-- CREATE TABLE IF NOT EXISTS branch (
--     branch_id SERIAL PRIMARY KEY,
--     branch_name VARCHAR,
--     created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
-- );

CREATE TABLE IF NOT EXISTS "user" (
    user_id SERIAL,
    first_name TEXT,
    last_name TEXT,
    username TEXT,
    password_hash TEXT, 
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    active BOOL DEFAULT FALSE,
    CONSTRAINT pk_user PRIMARY KEY (user_id)
);

-- CREATE TABLE IF NOT EXISTS app_role_group (
--     user_id INT NOT NULL,
--     app_role_id INT NOT NULL,
--     CONSTRAINT pk_app_role_group PRIMARY KEY (user_id,app_role_id)
-- );

CREATE TABLE IF NOT EXISTS user_canvas_group (
    user_id INT NOT NULL,
    canvas_id INT NOT NULL,
    CONSTRAINT pk_user_canvas_group PRIMARY KEY (user_id,canvas_id)
);

--! test
INSERT INTO canvas (canvas_name) VALUES ('dashboard 1');
INSERT INTO canvas (canvas_name) VALUES ('dashboard 2');
INSERT INTO canvas (canvas_name) VALUES ('dashboard 3');
INSERT INTO widget (widget_type_id, canvas_id, layout_data) VALUES (1, 1, '{"x": 0, "y": 0, "w": 6, "h":8}');
insert into user_canvas_group (user_id ,canvas_id) values(1,1);
--! test