CREATE TABLE IF NOT EXISTS widget (
    widget_id SERIAL,
    widget_type_id INT NOT NULL,
    canvas_id INT NOT NULL,
    device_id INT,
    widget_label TEXT,
    layout_data JSONB NOT NULL,
    widget_color JSONB,
    custom_chart_data JSONB,
    CONSTRAINT pk_widget PRIMARY KEY (widget_id)
);

CREATE INDEX IF NOT EXISTS ix_canvas ON widget (canvas_id);
CREATE INDEX IF NOT EXISTS ix_device ON widget (device_id);

CREATE TABLE IF NOT EXISTS widget_type (
    widget_type_id SERIAL,
    widget_type_name TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_widget_type PRIMARY KEY (widget_type_id),
    CONSTRAINT uq_widget_type UNIQUE (widget_type_name)
);

CREATE TABLE IF NOT EXISTS canvas (
    canvas_id SERIAL,
    canvas_name TEXT,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    CONSTRAINT pk_canvas PRIMARY KEY (canvas_id)
);

CREATE TABLE IF NOT EXISTS "user" (
    user_id SERIAL,
    first_name TEXT,
    last_name TEXT,
    username TEXT,
    password_hash TEXT, 
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    active BOOL DEFAULT FALSE,
    role_id INT,
    CONSTRAINT pk_user PRIMARY KEY (user_id)
);

CREATE TABLE IF NOT EXISTS device (
    device_id SERIAL,
    device_name TEXT NOT NULL,
    value_data INT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ DEFAULT NULL,
    is_active BOOL NOT NULL,
    is_connected BOOL DEFAULT FALSE,
    last_seen_at TIMESTAMPTZ,
    CONSTRAINT pk_device_id PRIMARY KEY (device_id)
);

CREATE UNIQUE INDEX uq_device_name_active
ON device (device_name) 
WHERE deleted_at IS NULL;

CREATE TABLE IF NOT EXISTS audit_log (
    id SERIAL,
    entity_type TEXT,
    entity_id TEXT,
    action TEXT,
    changed_by INT,
    old_data JSONB,
    new_data JSONB,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_audit_log_id PRIMARY KEY (id)
);

CREATE INDEX IF NOT EXISTS ix_audit_log_entity_type_id ON audit_log (entity_type,entity_id);
CREATE INDEX IF NOT EXISTS ix_audit_log_created_at ON audit_log (created_at);
CREATE INDEX IF NOT EXISTS ix_audit_log_action ON audit_log (action);

CREATE TABLE IF NOT EXISTS device_request_log (
    id SERIAL,
    device_id INT NOT NULL,
    value_data INT NOT NULL,
    action TEXT,
    origin TEXT,
    sent_at TIMESTAMPTZ,
    CONSTRAINT pk_device_request_log_id PRIMARY KEY (id)
);

CREATE TYPE schedule_type AS ENUM ('one_time', 'recurring');
CREATE TYPE schedule_status AS ENUM ('active', 'completed', 'cancelled');

CREATE TABLE schedule (
    schedule_id UUID DEFAULT gen_random_uuid(),
    device_id INT NOT NULL,
    action TEXT NOT NULL, 
    schedule_type schedule_type NOT NULL,
    status schedule_status NOT NULL DEFAULT 'active',
    start_time TIMESTAMPTZ NOT NULL, 
    end_time TIMESTAMPTZ, 
    cron_expression TEXT, 
    last_run_at TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_schedule_id PRIMARY KEY (schedule_id)
);

CREATE TABLE IF NOT EXISTS role (
    role_id SERIAL,
    role_name TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    CONSTRAINT pk_role_id PRIMARY KEY (role_id),
    CONSTRAINT uq_role_name UNIQUE (role_name)
);

CREATE TABLE IF NOT EXISTS menu (
    menu_id SERIAL,
    menu_name TEXT NOT NULL,
    parent_id INT DEFAULT NULL,
    CONSTRAINT pk_menu_id PRIMARY KEY (menu_id),
    CONSTRAINT uq_menu_name UNIQUE (menu_name),
    CONSTRAINT fk_menu_parent FOREIGN KEY (parent_id) REFERENCES menu(menu_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS action (
    action_id SERIAL,
    action_name TEXT NOT NULL,
    CONSTRAINT pk_action_id PRIMARY KEY (action_id),
    CONSTRAINT uq_action_name UNIQUE (action_name)
);

CREATE TABLE IF NOT EXISTS menu_action (
    menu_id INT NOT NULL,
    action_id INT NOT NULL,
    CONSTRAINT pk_menu_action PRIMARY KEY (menu_id, action_id),
    CONSTRAINT fk_ma_menu FOREIGN KEY (menu_id) REFERENCES menu(menu_id) ON DELETE CASCADE,
    CONSTRAINT fk_ma_action FOREIGN KEY (action_id) REFERENCES action(action_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS role_permission (
    role_id INT NOT NULL,
    menu_id INT NOT NULL,
    action_id INT NOT NULL,
    CONSTRAINT pk_role_permission PRIMARY KEY (role_id, menu_id, action_id),
    CONSTRAINT fk_rp_role FOREIGN KEY (role_id) REFERENCES role(role_id) ON DELETE CASCADE,
    CONSTRAINT fk_rp_valid_action FOREIGN KEY (menu_id, action_id) 
        REFERENCES menu_action(menu_id, action_id) ON DELETE CASCADE
);

CREATE TABLE IF NOT EXISTS canvas_role (
    canvas_id INT NOT NULL,
    role_id INT NOT NULL,
    CONSTRAINT pk_canvas_role PRIMARY KEY (canvas_id, role_id),
    CONSTRAINT fk_cr_canvas FOREIGN KEY (canvas_id) REFERENCES canvas(canvas_id) ON DELETE CASCADE,
    CONSTRAINT fk_cr_role FOREIGN KEY (role_id) REFERENCES role(role_id) ON DELETE CASCADE
);
